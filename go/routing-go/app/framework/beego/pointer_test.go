package main

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)
/*
https://studygolang.gitbook.io/learn-go-with-tests/go-ji-chu/pointers-and-errors
 */
type Bitcoin int

func (bitcoin Bitcoin) String() string {
	return fmt.Sprintf("%d Bitcoin", bitcoin)
}

type Stringer interface {
	String() string
}

type Wallet struct {
	balance Bitcoin
}

func (wallet *Wallet) Deposit(amount Bitcoin)  {
	fmt.Println("address of balance in Deposit is", &wallet.balance)
	wallet.balance += amount
}

func (wallet *Wallet) Balance() Bitcoin  {
	return wallet.balance
}

func (wallet *Wallet) Withdraw(amount Bitcoin) error  {
	if wallet.balance < amount {
		return errors.New("not enough money")
	}

	wallet.balance -= amount

	return nil
}

func TestWallet(test *testing.T)  {
	assertBalance := func(test *testing.T, wallet Wallet, want Bitcoin) {
		got := wallet.Balance()

		if got != want {
			test.Errorf("%#v got %d want %d", wallet, got, want)
		}
	}

	assertError := func(test *testing.T, err error) {
		if err == nil {
			test.Error("wanted an error but none")
		}
	}

	test.Run("Deposit", func(test *testing.T) {
		wallet := Wallet{}
		fmt.Println("address of balance in test is", &wallet.balance)
		wallet.Deposit(Bitcoin(10)) // 在 Go 中，当调用一个函数或方法时，参数会被复制
		assertBalance(test, wallet, Bitcoin(10))
	})

	test.Run("Withdraw withing balance", func(test *testing.T) {
		wallet := Wallet{balance: Bitcoin(20)}
		err := wallet.Withdraw(Bitcoin(10))
		assert.Nil(test, err)
		assertBalance(test, wallet, Bitcoin(10))
	})

	test.Run("Withdraw over balance", func(test *testing.T) {
		wallet := Wallet{balance: Bitcoin(20)}
		err := wallet.Withdraw(Bitcoin(100))
		assertError(test, err)
		assertBalance(test, wallet, Bitcoin(20))
	})
}

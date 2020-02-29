package jwe

import (
	"testing"
	"k8s.io/client-go/kubernetes/fake"
	)


func getKeyHolder() KeyHolder {
	client := fake.NewSimpleClientset()

}

func TestNewRSAKeyHolder(test *testing.T) {
	holder := getKeyHolder()
	if holder == nil {
		test.Fatalf("NewRSAKeyHolder(): Expected key holder not to be nil")
	}
}

func TestRsaKeyHolder_Encrypter(test *testing.T) {
}

func TestRsaKeyHolder_Key(test *testing.T) {
}



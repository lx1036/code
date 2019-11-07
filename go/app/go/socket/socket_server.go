package main

// use Go to create TCP server easily

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

const (
	Address = "localhost:8081"
)

func server() {
	fmt.Println("Launching server")
	listener, _ := net.Listen("tcp", Address)

	for {
		connection, _ := listener.Accept()
		message, _ := bufio.NewReader(connection).ReadString('\n')
		fmt.Printf("Received from client message: %s", message)
		_, _ = connection.Write([]byte(strings.ToUpper(message)))
	}
}

func client() {
	fmt.Println("Launching client")
	connection, _ := net.Dial("tcp", Address)

	for {
		text, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		fmt.Print(text)
		_, err := fmt.Fprintf(connection, text)
		if err != nil {
			log.Fatal(err)
		}
		message, _ := bufio.NewReader(connection).ReadString('\n')
		fmt.Printf("Received from server message: %s", message)
	}
}

func callFanyiSoCom() {
	query := "我想回家"
	url := "http://fanyi.so.com/index/search?auto_trans=zh_en&src=360AI&query=" + query
	response, err := http.Get(url)
	if err != nil {
		_ = fmt.Errorf("%v", err)
	}
	body, _ := ioutil.ReadAll(response.Body)
	fmt.Println(string(body))
}

func main() {
	/*go server()
	  go client()

	  select {}*/

	callFanyiSoCom()
}

package sdk

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

type Person struct {
	Name string `json:"name"`
}

type Student struct {
	Person
	Age int `json:"age"`
}

func TestNestedStruct(test *testing.T) {
	var student Student
	data := `
	{
		"name": "test",
		"age": 10
	}
`
	err := json.Unmarshal([]byte(data), &student)
	if err != nil {
		panic(err)
	}

	fmt.Println(student.Name, student.Age)
}

func TestCreateService(test *testing.T) {
	kongClient := NewClient(NewDefaultConfig(Config{}))
	service := kongClient.Services().Create(Service{
		Name:           "",
		Protocol:       "",
		Host:           "",
		Port:           0,
		Path:           "",
		Retries:        0,
		ConnectTimeout: 0,
		WriteTimeout:   0,
		ReadTimeout:    0,
		Tags:           nil,
		ClientCertificate: struct {
			Id string `json:"id" yaml:"id"`
		}{},
		TlsVerify:      false,
		TlsVerifyDepth: "",
		CreatedAt:      0,
		UpdatedAt:      0,
	})

	services := kongClient.Services().List(&ServiceQuery{})
	service := kongClient.Services().GetServiceById(service.Id)
	kongClient.Services().Update(Service{})
	err := kongClient.Services().Delete(service.Id)
	assert.Nil(test, err)
}

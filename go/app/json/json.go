package main

import (
	"encoding/json"
	"fmt"
)

type Message struct {
	Name string
	Body string
	Time int64
}

func main() {
	m := Message{
		Name: "Alice",
		Body: "Hello",
		Time: 1294706395881547000,
	}

	b, _ := json.Marshal(m)

	var m2 Message
	json.Unmarshal(b, &m2)

	//fmt.Println(b, string(b), m2.Name)

	bb := []byte(`{"NAme": "Alice2", "Age":6}, "Parents": ["parent1", "parent2"]"`)
	var people interface{}
	_ = json.Unmarshal(bb, &people)
	fmt.Println(people)

	type FamilyMember struct {
		Name    string `json:"name"`
		Age     int    `json:"age"`
		Parents []string
	}
	b3 := []byte(`{"Name":"Wednesday","Age":6,"Parents":["Gomez","Morticia"]}`)
	var m3 FamilyMember
	_ = json.Unmarshal(b3, &m3)
	fmt.Println(m3.Parents)
}

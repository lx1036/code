package main

/**
  https://blog.golang.org/json-and-go
*/
import (
	"encoding/json"
	"fmt"
)

type FamilyMember struct {
	Name    string
	Age     int
	Parents []string
}

func main() {
	var f = []byte(`{"NAme": "Alice2", "Age":6}, "Parents": ["parent1", "parent2"]"`)
	var m FamilyMember
	_ = json.Unmarshal(f, &m)
	fmt.Println(m.Parents)
}

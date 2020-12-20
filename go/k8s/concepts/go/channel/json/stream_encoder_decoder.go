package main

import (
	"encoding/json"
	"log"
	"os"
)

func main() {
	dec := json.NewDecoder(os.Stdin)
	enc := json.NewEncoder(os.Stdout)

	for {
		var v map[string]interface{}
		if err := dec.Decode(&v); err != nil {
			log.Println(err)
			return
		}
		for key := range v {
			if key != "Name" {
				delete(v, key)
			}
		}
		if err := enc.Encode(v); err != nil {
			log.Println(err)
		}
	}
}

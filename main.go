package main

import (
	"fmt"
	"encoding/json"
	"os"
)

func main() {

	res, err := Request("GET", "https://www.google.com/", nil, nil)

	if err != nil {
		fmt.Println("ERROR URL:", err)
	}

	enc := json.NewEncoder(os.Stderr)
	enc.SetIndent("", "  ")
	enc.Encode(res.Stats())
}
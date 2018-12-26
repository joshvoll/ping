package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	// get args
	addr := os.Args[1]
	res, err := Request("GET", addr, nil, nil)

	if err != nil {
		fmt.Println("ERROR URL:", err)
	}

	enc := json.NewEncoder(os.Stderr)
	enc.SetIndent("", "  ")
	enc.Encode(res.Stats())
}

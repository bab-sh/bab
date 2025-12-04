//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bab-sh/bab/internal/babfile"
)

func main() {
	s := babfile.GenerateSchema()

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling schema: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(data))
}

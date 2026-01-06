package main

import (
	"fmt"
	"os"

	"github.com/podoru/podoru-chain/internal/crypto"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <key-file>\n", os.Args[0])
		os.Exit(1)
	}

	keyPath := os.Args[1]
	privateKey, err := crypto.LoadPrivateKeyFromFile(keyPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading private key: %v\n", err)
		os.Exit(1)
	}

	address, err := crypto.AddressFromPrivateKey(privateKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error deriving address: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Address: %s\n", address)
}

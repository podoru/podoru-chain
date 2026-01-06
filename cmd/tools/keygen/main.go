package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/podoru/podoru-chain/internal/crypto"
)

func main() {
	outputPath := flag.String("output", "", "Output path for the private key file")
	showAddress := flag.Bool("address", true, "Show the derived address")
	flag.Parse()

	// Generate key pair
	privateKey, err := crypto.GenerateKeyPair()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating key pair: %v\n", err)
		os.Exit(1)
	}

	// Get address
	address, err := crypto.AddressFromPrivateKey(privateKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error deriving address: %v\n", err)
		os.Exit(1)
	}

	// Save or print key
	if *outputPath != "" {
		// Create directory if it doesn't exist
		dir := filepath.Dir(*outputPath)
		if err := os.MkdirAll(dir, 0700); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
			os.Exit(1)
		}

		// Save private key
		if err := crypto.SavePrivateKeyToFile(privateKey, *outputPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving private key: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Private key saved to: %s\n", *outputPath)
	} else {
		// Print private key in hex
		keyBytes := crypto.PrivateKeyToBytes(privateKey)
		fmt.Printf("Private Key: %x\n", keyBytes)
	}

	// Show address
	if *showAddress {
		fmt.Printf("Address: %s\n", address)
	}

	// Show public key
	publicKey := crypto.GetPublicKey(privateKey)
	publicKeyBytes := crypto.PublicKeyToBytes(publicKey)
	fmt.Printf("Public Key: %x\n", publicKeyBytes)
}

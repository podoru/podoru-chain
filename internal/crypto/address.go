package crypto

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
)

// AddressFromPublicKey derives an Ethereum-style address from a public key
// Address is the last 20 bytes of the Keccak256 hash of the public key
func AddressFromPublicKey(publicKey *ecdsa.PublicKey) (string, error) {
	if publicKey == nil {
		return "", errors.New("public key is nil")
	}

	address := crypto.PubkeyToAddress(*publicKey)
	return address.Hex(), nil
}

// AddressFromPrivateKey derives an address from a private key
func AddressFromPrivateKey(privateKey *ecdsa.PrivateKey) (string, error) {
	if err := ValidatePrivateKey(privateKey); err != nil {
		return "", err
	}

	publicKey := GetPublicKey(privateKey)
	return AddressFromPublicKey(publicKey)
}

// IsValidAddress checks if a string is a valid Ethereum-style address
func IsValidAddress(address string) bool {
	// Address should start with 0x and be 42 characters long (0x + 40 hex chars)
	if !strings.HasPrefix(address, "0x") {
		return false
	}

	if len(address) != 42 {
		return false
	}

	// Check if the rest is valid hex
	_, err := hex.DecodeString(address[2:])
	return err == nil
}

// NormalizeAddress ensures an address is in proper format
func NormalizeAddress(address string) string {
	address = strings.ToLower(strings.TrimSpace(address))
	if !strings.HasPrefix(address, "0x") {
		address = "0x" + address
	}
	return address
}

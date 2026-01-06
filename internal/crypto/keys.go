package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
)

// GenerateKeyPair generates a new ECDSA key pair
func GenerateKeyPair() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

// PrivateKeyToBytes converts a private key to bytes
func PrivateKeyToBytes(privateKey *ecdsa.PrivateKey) []byte {
	return crypto.FromECDSA(privateKey)
}

// PrivateKeyFromBytes converts bytes to a private key
func PrivateKeyFromBytes(b []byte) (*ecdsa.PrivateKey, error) {
	return crypto.ToECDSA(b)
}

// PublicKeyToBytes converts a public key to bytes
func PublicKeyToBytes(publicKey *ecdsa.PublicKey) []byte {
	return crypto.FromECDSAPub(publicKey)
}

// PublicKeyFromBytes converts bytes to a public key
func PublicKeyFromBytes(b []byte) (*ecdsa.PublicKey, error) {
	return crypto.UnmarshalPubkey(b)
}

// SavePrivateKeyToFile saves a private key to a file
func SavePrivateKeyToFile(privateKey *ecdsa.PrivateKey, filePath string) error {
	keyBytes := PrivateKeyToBytes(privateKey)
	keyHex := hex.EncodeToString(keyBytes)
	return os.WriteFile(filePath, []byte(keyHex), 0600)
}

// LoadPrivateKeyFromFile loads a private key from a file
func LoadPrivateKeyFromFile(filePath string) (*ecdsa.PrivateKey, error) {
	keyHex, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	keyBytes, err := hex.DecodeString(string(keyHex))
	if err != nil {
		return nil, fmt.Errorf("failed to decode key hex: %w", err)
	}

	return PrivateKeyFromBytes(keyBytes)
}

// GetPublicKey returns the public key from a private key
func GetPublicKey(privateKey *ecdsa.PrivateKey) *ecdsa.PublicKey {
	if privateKey == nil {
		return nil
	}
	return &privateKey.PublicKey
}

// ValidatePrivateKey validates that a private key is valid
func ValidatePrivateKey(privateKey *ecdsa.PrivateKey) error {
	if privateKey == nil {
		return errors.New("private key is nil")
	}
	if privateKey.D == nil {
		return errors.New("private key D is nil")
	}
	if privateKey.PublicKey.X == nil || privateKey.PublicKey.Y == nil {
		return errors.New("public key coordinates are nil")
	}
	return nil
}

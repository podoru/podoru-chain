package crypto

import (
	"crypto/ecdsa"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
)

// Sign signs a hash with a private key
func Sign(hash []byte, privateKey *ecdsa.PrivateKey) ([]byte, error) {
	if err := ValidatePrivateKey(privateKey); err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	if len(hash) != 32 {
		return nil, errors.New("hash must be 32 bytes")
	}

	signature, err := crypto.Sign(hash, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %w", err)
	}

	return signature, nil
}

// Verify verifies a signature against a hash and public key
func Verify(hash []byte, signature []byte, publicKey *ecdsa.PublicKey) bool {
	if publicKey == nil {
		return false
	}

	if len(hash) != 32 {
		return false
	}

	if len(signature) != 65 {
		return false
	}

	// Recover the public key from signature
	recoveredPub, err := crypto.SigToPub(hash, signature)
	if err != nil {
		return false
	}

	// Compare recovered public key with provided public key
	recoveredAddress := crypto.PubkeyToAddress(*recoveredPub)
	providedAddress := crypto.PubkeyToAddress(*publicKey)

	return recoveredAddress == providedAddress
}

// RecoverPublicKey recovers the public key from a signature and hash
func RecoverPublicKey(hash []byte, signature []byte) (*ecdsa.PublicKey, error) {
	if len(hash) != 32 {
		return nil, errors.New("hash must be 32 bytes")
	}

	if len(signature) != 65 {
		return nil, errors.New("signature must be 65 bytes")
	}

	publicKey, err := crypto.SigToPub(hash, signature)
	if err != nil {
		return nil, fmt.Errorf("failed to recover public key: %w", err)
	}

	return publicKey, nil
}

// RecoverAddress recovers the address from a signature and hash
func RecoverAddress(hash []byte, signature []byte) (string, error) {
	publicKey, err := RecoverPublicKey(hash, signature)
	if err != nil {
		return "", err
	}

	return AddressFromPublicKey(publicKey)
}

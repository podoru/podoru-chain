package blockchain

import (
	"errors"
	"math/big"
)

const (
	// DefaultBaseFee is the default minimum fee per transaction (in wei)
	DefaultBaseFee = 1000

	// DefaultPerByteFee is the default fee per byte of transaction data (in wei)
	DefaultPerByteFee = 10
)

// GasConfig holds gas-related configuration
type GasConfig struct {
	BaseFee    *big.Int // Minimum fee per transaction
	PerByteFee *big.Int // Fee per byte of transaction data
}

// GasConfigJSON is the JSON representation of GasConfig
type GasConfigJSON struct {
	BaseFee    string `json:"base_fee"`
	PerByteFee string `json:"per_byte_fee"`
}

// DefaultGasConfig returns the default gas configuration
func DefaultGasConfig() *GasConfig {
	return &GasConfig{
		BaseFee:    big.NewInt(DefaultBaseFee),
		PerByteFee: big.NewInt(DefaultPerByteFee),
	}
}

// NewGasConfig creates a gas config from base fee and per-byte fee
func NewGasConfig(baseFee, perByteFee *big.Int) *GasConfig {
	if baseFee == nil {
		baseFee = big.NewInt(DefaultBaseFee)
	}
	if perByteFee == nil {
		perByteFee = big.NewInt(DefaultPerByteFee)
	}
	return &GasConfig{
		BaseFee:    baseFee,
		PerByteFee: perByteFee,
	}
}

// GasConfigFromJSON creates a GasConfig from JSON representation
func GasConfigFromJSON(json *GasConfigJSON) (*GasConfig, error) {
	if json == nil {
		return DefaultGasConfig(), nil
	}

	baseFee := big.NewInt(DefaultBaseFee)
	if json.BaseFee != "" {
		var ok bool
		baseFee, ok = new(big.Int).SetString(json.BaseFee, 10)
		if !ok {
			return nil, errors.New("invalid base_fee")
		}
	}

	perByteFee := big.NewInt(DefaultPerByteFee)
	if json.PerByteFee != "" {
		var ok bool
		perByteFee, ok = new(big.Int).SetString(json.PerByteFee, 10)
		if !ok {
			return nil, errors.New("invalid per_byte_fee")
		}
	}

	return &GasConfig{
		BaseFee:    baseFee,
		PerByteFee: perByteFee,
	}, nil
}

// ToJSON converts GasConfig to JSON representation
func (gc *GasConfig) ToJSON() *GasConfigJSON {
	return &GasConfigJSON{
		BaseFee:    gc.BaseFee.String(),
		PerByteFee: gc.PerByteFee.String(),
	}
}

// CalculateGasFee calculates the gas fee for a transaction of given size
// Formula: baseFee + (txSize * perByteFee)
func (gc *GasConfig) CalculateGasFee(txSize int) *big.Int {
	if txSize < 0 {
		txSize = 0
	}

	sizeFee := new(big.Int).Mul(gc.PerByteFee, big.NewInt(int64(txSize)))
	totalFee := new(big.Int).Add(gc.BaseFee, sizeFee)

	return totalFee
}

// Validate validates the gas configuration
func (gc *GasConfig) Validate() error {
	if gc.BaseFee == nil {
		return errors.New("base_fee is required")
	}
	if gc.BaseFee.Sign() < 0 {
		return errors.New("base_fee cannot be negative")
	}
	if gc.PerByteFee == nil {
		return errors.New("per_byte_fee is required")
	}
	if gc.PerByteFee.Sign() < 0 {
		return errors.New("per_byte_fee cannot be negative")
	}
	return nil
}

// Clone creates a copy of the gas config
func (gc *GasConfig) Clone() *GasConfig {
	return &GasConfig{
		BaseFee:    new(big.Int).Set(gc.BaseFee),
		PerByteFee: new(big.Int).Set(gc.PerByteFee),
	}
}

// IsZeroFee returns true if gas fees are effectively disabled
func (gc *GasConfig) IsZeroFee() bool {
	return gc.BaseFee.Sign() == 0 && gc.PerByteFee.Sign() == 0
}

// GasEstimate represents a gas fee estimate
type GasEstimate struct {
	TransactionSize int      `json:"transaction_size"`
	BaseFee         *big.Int `json:"base_fee"`
	SizeFee         *big.Int `json:"size_fee"`
	TotalFee        *big.Int `json:"total_fee"`
}

// EstimateGas creates a gas estimate for a transaction size
func (gc *GasConfig) EstimateGas(txSize int) *GasEstimate {
	if txSize < 0 {
		txSize = 0
	}

	sizeFee := new(big.Int).Mul(gc.PerByteFee, big.NewInt(int64(txSize)))
	totalFee := new(big.Int).Add(gc.BaseFee, sizeFee)

	return &GasEstimate{
		TransactionSize: txSize,
		BaseFee:         new(big.Int).Set(gc.BaseFee),
		SizeFee:         sizeFee,
		TotalFee:        totalFee,
	}
}

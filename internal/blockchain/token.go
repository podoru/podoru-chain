package blockchain

import (
	"errors"
	"math/big"
	"strings"
)

const (
	// TokenDecimals is the number of decimals for PDR token
	TokenDecimals = 18

	// BalanceKeyPrefix is the prefix for balance storage keys
	BalanceKeyPrefix = "balance:"

	// TokenName is the name of the native token
	TokenName = "Podoru"

	// TokenSymbol is the symbol of the native token
	TokenSymbol = "PDR"

	// InitialSupplyString is 100 million PDR in wei (100_000_000 * 10^18)
	InitialSupplyString = "100000000000000000000000000"
)

var (
	// InitialSupply as big.Int
	InitialSupply = func() *big.Int {
		supply, _ := new(big.Int).SetString(InitialSupplyString, 10)
		return supply
	}()

	// OnePDR is one PDR in wei (10^18)
	OnePDR = new(big.Int).Exp(big.NewInt(10), big.NewInt(TokenDecimals), nil)

	// ZeroBalance represents zero balance
	ZeroBalance = big.NewInt(0)
)

// Balance represents an account balance using big.Int
type Balance struct {
	Amount *big.Int
}

// NewBalance creates a new balance
func NewBalance(amount *big.Int) *Balance {
	if amount == nil {
		amount = big.NewInt(0)
	}
	return &Balance{Amount: new(big.Int).Set(amount)}
}

// NewBalanceFromString creates a balance from a string
func NewBalanceFromString(s string) (*Balance, error) {
	amount, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return nil, errors.New("invalid balance string")
	}
	return NewBalance(amount), nil
}

// BalanceFromBytes deserializes a balance from bytes
func BalanceFromBytes(data []byte) (*Balance, error) {
	if len(data) == 0 {
		return NewBalance(big.NewInt(0)), nil
	}
	if len(data) > 32 {
		return nil, errors.New("invalid balance data: too long")
	}
	amount := new(big.Int).SetBytes(data)
	return NewBalance(amount), nil
}

// ToBytes serializes the balance to bytes (big-endian, up to 32 bytes)
func (b *Balance) ToBytes() []byte {
	if b.Amount == nil || b.Amount.Sign() == 0 {
		return []byte{}
	}
	return b.Amount.Bytes()
}

// Add adds amount to balance
func (b *Balance) Add(amount *big.Int) {
	if amount == nil {
		return
	}
	b.Amount.Add(b.Amount, amount)
}

// Sub subtracts amount from balance, returns error if insufficient
func (b *Balance) Sub(amount *big.Int) error {
	if amount == nil {
		return nil
	}
	if b.Amount.Cmp(amount) < 0 {
		return errors.New("insufficient balance")
	}
	b.Amount.Sub(b.Amount, amount)
	return nil
}

// Cmp compares balance with another amount
// Returns -1 if b < amount, 0 if equal, 1 if b > amount
func (b *Balance) Cmp(amount *big.Int) int {
	if amount == nil {
		amount = big.NewInt(0)
	}
	return b.Amount.Cmp(amount)
}

// IsZero returns true if balance is zero
func (b *Balance) IsZero() bool {
	return b.Amount.Sign() == 0
}

// String returns the balance as a string
func (b *Balance) String() string {
	return b.Amount.String()
}

// Clone creates a copy of the balance
func (b *Balance) Clone() *Balance {
	return NewBalance(b.Amount)
}

// BalanceKey returns the state key for an address's balance
func BalanceKey(address string) string {
	return BalanceKeyPrefix + strings.ToLower(address)
}

// IsBalanceKey checks if a key is a balance key
func IsBalanceKey(key string) bool {
	return strings.HasPrefix(key, BalanceKeyPrefix)
}

// AddressFromBalanceKey extracts the address from a balance key
func AddressFromBalanceKey(key string) string {
	if !IsBalanceKey(key) {
		return ""
	}
	return key[len(BalanceKeyPrefix):]
}

// TokenConfig holds token configuration from genesis
type TokenConfig struct {
	Name          string `json:"name"`
	Symbol        string `json:"symbol"`
	Decimals      int    `json:"decimals"`
	InitialSupply string `json:"initial_supply"`
}

// DefaultTokenConfig returns the default token configuration
func DefaultTokenConfig() *TokenConfig {
	return &TokenConfig{
		Name:          TokenName,
		Symbol:        TokenSymbol,
		Decimals:      TokenDecimals,
		InitialSupply: InitialSupplyString,
	}
}

// Validate validates the token configuration
func (tc *TokenConfig) Validate() error {
	if tc.Name == "" {
		return errors.New("token name is required")
	}
	if tc.Symbol == "" {
		return errors.New("token symbol is required")
	}
	if tc.Decimals < 0 || tc.Decimals > 18 {
		return errors.New("decimals must be between 0 and 18")
	}
	if tc.InitialSupply != "" {
		_, ok := new(big.Int).SetString(tc.InitialSupply, 10)
		if !ok {
			return errors.New("invalid initial supply")
		}
	}
	return nil
}

// GetInitialSupply returns the initial supply as big.Int
func (tc *TokenConfig) GetInitialSupply() *big.Int {
	if tc.InitialSupply == "" {
		return InitialSupply
	}
	supply, ok := new(big.Int).SetString(tc.InitialSupply, 10)
	if !ok {
		return InitialSupply
	}
	return supply
}

// FormatBalance formats a balance in wei to a human-readable string
func FormatBalance(weiAmount *big.Int) string {
	if weiAmount == nil || weiAmount.Sign() == 0 {
		return "0 PDR"
	}

	// Convert to float for display
	weiFloat := new(big.Float).SetInt(weiAmount)
	oneToken := new(big.Float).SetInt(OnePDR)
	result := new(big.Float).Quo(weiFloat, oneToken)

	return result.Text('f', 6) + " PDR"
}

// ParsePDR converts a PDR amount string to wei
func ParsePDR(pdrAmount string) (*big.Int, error) {
	// Parse the amount as a float
	amount, _, err := big.ParseFloat(pdrAmount, 10, 256, big.ToNearestEven)
	if err != nil {
		return nil, err
	}

	// Multiply by 10^18
	oneToken := new(big.Float).SetInt(OnePDR)
	weiFloat := new(big.Float).Mul(amount, oneToken)

	// Convert to int
	wei, _ := weiFloat.Int(nil)
	return wei, nil
}

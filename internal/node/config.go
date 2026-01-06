package node

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

// Config holds node configuration
type Config struct {
	// Node identity
	NodeType   NodeType `mapstructure:"node_type"`
	Address    string   `mapstructure:"address"`
	PrivateKey string   `mapstructure:"private_key"`

	// Network
	P2PPort        int      `mapstructure:"p2p_port"`
	P2PBindAddr    string   `mapstructure:"p2p_bind_addr"`
	BootstrapPeers []string `mapstructure:"bootstrap_peers"`
	MaxPeers       int      `mapstructure:"max_peers"`

	// API
	APIEnabled  bool   `mapstructure:"api_enabled"`
	APIPort     int    `mapstructure:"api_port"`
	APIBindAddr string `mapstructure:"api_bind_addr"`

	// Storage
	DataDir string `mapstructure:"data_dir"`

	// Consensus
	Authorities []string      `mapstructure:"authorities"`
	BlockTime   time.Duration `mapstructure:"block_time"`

	// Genesis
	GenesisPath string `mapstructure:"genesis_path"`
}

// LoadConfig loads configuration from a file
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	// Set default values
	v.SetDefault("node_type", "full")
	v.SetDefault("p2p_port", 9000)
	v.SetDefault("p2p_bind_addr", "0.0.0.0")
	v.SetDefault("max_peers", 50)
	v.SetDefault("api_enabled", true)
	v.SetDefault("api_port", 8545)
	v.SetDefault("api_bind_addr", "0.0.0.0")
	v.SetDefault("data_dir", "./data")
	v.SetDefault("block_time", "5s")

	// Read config file
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate node type
	if !c.NodeType.IsValid() {
		return fmt.Errorf("invalid node type: %s", c.NodeType)
	}

	// For producer nodes, address and private key are required
	if c.NodeType == NodeTypeProducer {
		if c.Address == "" {
			return errors.New("address is required for producer nodes")
		}
		if c.PrivateKey == "" {
			return errors.New("private_key is required for producer nodes")
		}

		// Check if private key file exists
		if _, err := os.Stat(c.PrivateKey); os.IsNotExist(err) {
			return fmt.Errorf("private key file not found: %s", c.PrivateKey)
		}
	}

	// Validate ports
	if c.P2PPort <= 0 || c.P2PPort > 65535 {
		return fmt.Errorf("invalid p2p_port: %d", c.P2PPort)
	}

	if c.APIEnabled {
		if c.APIPort <= 0 || c.APIPort > 65535 {
			return fmt.Errorf("invalid api_port: %d", c.APIPort)
		}
	}

	// Validate authorities
	if len(c.Authorities) == 0 {
		return errors.New("no authorities specified")
	}

	// Validate genesis path
	if c.GenesisPath == "" {
		return errors.New("genesis_path is required")
	}

	// Check if genesis file exists
	if _, err := os.Stat(c.GenesisPath); os.IsNotExist(err) {
		return fmt.Errorf("genesis file not found: %s", c.GenesisPath)
	}

	// Validate block time
	if c.BlockTime <= 0 {
		return errors.New("block_time must be positive")
	}

	return nil
}

// IsProducer returns true if this is a producer node
func (c *Config) IsProducer() bool {
	return c.NodeType == NodeTypeProducer
}

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/podoru/podoru-chain/internal/api/rest"
	"github.com/podoru/podoru-chain/internal/node"
	"github.com/sirupsen/logrus"
)

var (
	configPath = flag.String("config", "", "Path to configuration file")
	version    = "1.0.0"
)

func main() {
	flag.Parse()

	// Setup logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Print banner
	printBanner()

	// Check config path
	if *configPath == "" {
		logger.Fatal("Config file path is required (use -config flag)")
	}

	// Load configuration
	logger.Infof("Loading configuration from %s...", *configPath)
	config, err := node.LoadConfig(*configPath)
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	// Create node
	logger.Info("Creating blockchain node...")
	n, err := node.NewNode(config)
	if err != nil {
		logger.Fatalf("Failed to create node: %v", err)
	}

	// Start node
	if err := n.Start(); err != nil {
		logger.Fatalf("Failed to start node: %v", err)
	}

	// Start API server if enabled
	var apiServer *rest.Server
	if config.APIEnabled {
		logger.Info("Starting REST API server...")
		apiServer = rest.NewServer(n, config.APIBindAddr, config.APIPort, logger)
		if err := apiServer.Start(); err != nil {
			logger.Fatalf("Failed to start API server: %v", err)
		}
	}

	logger.Info("Podoru Chain node is running")
	logger.Infof("Press Ctrl+C to stop")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down...")

	// Stop API server
	if apiServer != nil {
		if err := apiServer.Stop(); err != nil {
			logger.Errorf("Error stopping API server: %v", err)
		}
	}

	// Stop node
	if err := n.Stop(); err != nil {
		logger.Errorf("Error stopping node: %v", err)
	}

	logger.Info("Goodbye!")
}

func printBanner() {
	banner := `
╔═══════════════════════════════════════╗
║                                       ║
║        PODORU CHAIN v` + version + `         ║
║   Decentralized Blockchain Platform   ║
║                                       ║
╚═══════════════════════════════════════╝
`
	fmt.Println(banner)
}

package rest

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/podoru/podoru-chain/internal/blockchain"
	"github.com/podoru/podoru-chain/internal/crypto"
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error response
func writeError(w http.ResponseWriter, status int, err string) {
	writeJSON(w, status, Response{
		Success: false,
		Error:   err,
	})
}

// writeSuccess writes a success response
func writeSuccess(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// handleGetChainInfo returns blockchain information
func (s *Server) handleGetChainInfo(w http.ResponseWriter, r *http.Request) {
	info, err := s.node.GetChain().GetChainInfo()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSuccess(w, info)
}

// handleGetBlockByHash returns a block by its hash
func (s *Server) handleGetBlockByHash(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hashStr := vars["hash"]

	// Remove 0x prefix if present
	if len(hashStr) > 2 && hashStr[:2] == "0x" {
		hashStr = hashStr[2:]
	}

	hash, err := hex.DecodeString(hashStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid hash format")
		return
	}

	block, err := s.node.GetChain().GetBlockByHash(hash)
	if err != nil {
		writeError(w, http.StatusNotFound, "block not found")
		return
	}

	writeSuccess(w, block)
}

// handleGetBlockByHeight returns a block by its height
func (s *Server) handleGetBlockByHeight(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	heightStr := vars["height"]

	height, err := strconv.ParseUint(heightStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid height format")
		return
	}

	block, err := s.node.GetChain().GetBlockByHeight(height)
	if err != nil {
		writeError(w, http.StatusNotFound, "block not found")
		return
	}

	writeSuccess(w, block)
}

// handleGetLatestBlock returns the latest block
func (s *Server) handleGetLatestBlock(w http.ResponseWriter, r *http.Request) {
	block := s.node.GetChain().GetCurrentBlock()
	writeSuccess(w, block)
}

// handleGetTransaction returns a transaction by hash
func (s *Server) handleGetTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hashStr := vars["hash"]

	// Remove 0x prefix if present
	if len(hashStr) > 2 && hashStr[:2] == "0x" {
		hashStr = hashStr[2:]
	}

	hash, err := hex.DecodeString(hashStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid hash format")
		return
	}

	tx, err := s.node.GetChain().GetTransaction(hash)
	if err != nil {
		writeError(w, http.StatusNotFound, "transaction not found")
		return
	}

	writeSuccess(w, tx)
}

// SubmitTransactionRequest represents a transaction submission request
type SubmitTransactionRequest struct {
	Transaction *blockchain.Transaction `json:"transaction"`
}

// handleSubmitTransaction submits a new transaction
func (s *Server) handleSubmitTransaction(w http.ResponseWriter, r *http.Request) {
	var req SubmitTransactionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Transaction == nil {
		writeError(w, http.StatusBadRequest, "transaction is required")
		return
	}

	if err := s.node.SubmitTransaction(req.Transaction); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, map[string]string{
		"transaction_hash": fmt.Sprintf("0x%x", req.Transaction.ID),
		"status":           "submitted",
	})
}

// handleGetState returns a state value by key
func (s *Server) handleGetState(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	value, err := s.node.GetChain().GetState(key)
	if err != nil {
		writeError(w, http.StatusNotFound, "key not found")
		return
	}

	writeSuccess(w, map[string]interface{}{
		"key":   key,
		"value": value,
	})
}

// NodeInfo represents node information
type NodeInfo struct {
	Version string `json:"version"`
	Type    string `json:"type"`
	Address string `json:"address"`
	Peers   int    `json:"peers"`
}

// handleGetNodeInfo returns node information
func (s *Server) handleGetNodeInfo(w http.ResponseWriter, r *http.Request) {
	info := NodeInfo{
		Version: "1.0.0",
		Type:    "podoru-chain",
		Peers:   s.node.GetP2PServer().PeerCount(),
	}

	writeSuccess(w, info)
}

// handleGetPeers returns connected peers
func (s *Server) handleGetPeers(w http.ResponseWriter, r *http.Request) {
	peers := s.node.GetP2PServer().GetPeers()

	peerInfo := make([]map[string]string, len(peers))
	for i, peer := range peers {
		peerInfo[i] = map[string]string{
			"id":      peer.ID,
			"address": peer.Address,
		}
	}

	writeSuccess(w, peerInfo)
}

// handleHealthCheck returns node health status
func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	writeSuccess(w, map[string]string{
		"status": "healthy",
	})
}

// handleGetMempool returns pending transactions in mempool
func (s *Server) handleGetMempool(w http.ResponseWriter, r *http.Request) {
	transactions := s.node.GetMempool().GetAllPendingTransactions()

	writeSuccess(w, map[string]interface{}{
		"count":        len(transactions),
		"transactions": transactions,
	})
}

// BalanceResponse represents a balance response
type BalanceResponse struct {
	Address          string `json:"address"`
	Balance          string `json:"balance"`
	BalanceFormatted string `json:"balance_formatted"`
}

// handleGetBalance returns the balance for an address
func (s *Server) handleGetBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]

	// Validate address format
	if !crypto.IsValidAddress(address) {
		writeError(w, http.StatusBadRequest, "invalid address format")
		return
	}

	balance, err := s.node.GetChain().GetBalance(address)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSuccess(w, BalanceResponse{
		Address:          address,
		Balance:          balance.String(),
		BalanceFormatted: blockchain.FormatBalance(balance),
	})
}

// TokenInfoResponse represents token information
type TokenInfoResponse struct {
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Decimals    int    `json:"decimals"`
	TotalSupply string `json:"total_supply,omitempty"`
}

// handleGetTokenInfo returns token information
func (s *Server) handleGetTokenInfo(w http.ResponseWriter, r *http.Request) {
	chain := s.node.GetChain()
	tokenConfig := chain.GetTokenConfig()

	if tokenConfig == nil {
		// Return default token info for legacy chains
		writeSuccess(w, TokenInfoResponse{
			Name:     blockchain.TokenName,
			Symbol:   blockchain.TokenSymbol,
			Decimals: blockchain.TokenDecimals,
		})
		return
	}

	writeSuccess(w, TokenInfoResponse{
		Name:        tokenConfig.Name,
		Symbol:      tokenConfig.Symbol,
		Decimals:    tokenConfig.Decimals,
		TotalSupply: tokenConfig.InitialSupply,
	})
}

// GasEstimateRequest represents a gas estimate request
type GasEstimateRequest struct {
	TransactionSize int `json:"transaction_size"`
}

// GasEstimateResponse represents a gas estimate response
type GasEstimateResponse struct {
	TransactionSize int    `json:"transaction_size"`
	BaseFee         string `json:"base_fee"`
	PerByteFee      string `json:"per_byte_fee"`
	SizeFee         string `json:"size_fee"`
	TotalFee        string `json:"total_fee"`
	TotalFeeFormatted string `json:"total_fee_formatted"`
}

// handleEstimateGas estimates gas fee for a transaction
func (s *Server) handleEstimateGas(w http.ResponseWriter, r *http.Request) {
	var req GasEstimateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.TransactionSize <= 0 {
		writeError(w, http.StatusBadRequest, "transaction_size must be positive")
		return
	}

	chain := s.node.GetChain()
	estimate := chain.EstimateGasFee(req.TransactionSize)

	writeSuccess(w, GasEstimateResponse{
		TransactionSize:   estimate.TransactionSize,
		BaseFee:           estimate.BaseFee.String(),
		PerByteFee:        chain.GetGasConfig().PerByteFee.String(),
		SizeFee:           estimate.SizeFee.String(),
		TotalFee:          estimate.TotalFee.String(),
		TotalFeeFormatted: blockchain.FormatBalance(estimate.TotalFee),
	})
}

// GasConfigResponse represents gas configuration
type GasConfigResponse struct {
	Enabled    bool   `json:"enabled"`
	BaseFee    string `json:"base_fee"`
	PerByteFee string `json:"per_byte_fee"`
}

// handleGetGasConfig returns gas configuration
func (s *Server) handleGetGasConfig(w http.ResponseWriter, r *http.Request) {
	chain := s.node.GetChain()
	gasConfig := chain.GetGasConfig()

	if gasConfig == nil {
		writeSuccess(w, GasConfigResponse{
			Enabled:    false,
			BaseFee:    "0",
			PerByteFee: "0",
		})
		return
	}

	writeSuccess(w, GasConfigResponse{
		Enabled:    !gasConfig.IsZeroFee(),
		BaseFee:    gasConfig.BaseFee.String(),
		PerByteFee: gasConfig.PerByteFee.String(),
	})
}

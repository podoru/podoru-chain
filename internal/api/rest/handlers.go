package rest

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/podoru/podoru-chain/internal/blockchain"
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

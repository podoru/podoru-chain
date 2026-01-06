package rest

import (
	"encoding/json"
	"net/http"
)

// BatchStateRequest represents a batch state query request
type BatchStateRequest struct {
	Keys []string `json:"keys"`
}

// BatchStateResponse represents a batch state query response
type BatchStateResponse struct {
	Results map[string]interface{} `json:"results"`
}

// handleBatchGetState retrieves multiple state values at once
func (s *Server) handleBatchGetState(w http.ResponseWriter, r *http.Request) {
	var req BatchStateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.Keys) == 0 {
		writeError(w, http.StatusBadRequest, "keys array is required")
		return
	}

	// Limit batch size to prevent abuse
	if len(req.Keys) > 100 {
		writeError(w, http.StatusBadRequest, "maximum 100 keys per batch request")
		return
	}

	results := make(map[string]interface{})

	for _, key := range req.Keys {
		value, err := s.node.GetChain().GetState(key)
		if err != nil {
			// Key not found, return null
			results[key] = nil
		} else {
			results[key] = value
		}
	}

	writeSuccess(w, results)
}

// PrefixQueryRequest represents a prefix query request
type PrefixQueryRequest struct {
	Prefix string `json:"prefix"`
	Limit  int    `json:"limit,omitempty"`
}

// handleQueryByPrefix queries all keys with a given prefix
// Example: prefix "user:alice:" returns all alice's data
func (s *Server) handleQueryByPrefix(w http.ResponseWriter, r *http.Request) {
	var req PrefixQueryRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Prefix == "" {
		writeError(w, http.StatusBadRequest, "prefix is required")
		return
	}

	// Default limit
	if req.Limit == 0 || req.Limit > 1000 {
		req.Limit = 100
	}

	results, err := s.node.GetChain().QueryStateByPrefix(req.Prefix, req.Limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeSuccess(w, map[string]interface{}{
		"prefix":  req.Prefix,
		"count":   len(results),
		"results": results,
	})
}

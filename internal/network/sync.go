package network

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/podoru/podoru-chain/internal/blockchain"
	"github.com/sirupsen/logrus"
)

// Syncer handles blockchain synchronization
type Syncer struct {
	chain      *blockchain.Chain
	p2pServer  *P2PServer
	logger     *logrus.Logger
	isSyncing  bool
	syncPeriod time.Duration
}

// NewSyncer creates a new syncer
func NewSyncer(chain *blockchain.Chain, p2pServer *P2PServer, logger *logrus.Logger) *Syncer {
	if logger == nil {
		logger = logrus.New()
	}

	return &Syncer{
		chain:      chain,
		p2pServer:  p2pServer,
		logger:     logger,
		syncPeriod: 30 * time.Second,
	}
}

// SyncWithPeers synchronizes the blockchain with peers
func (s *Syncer) SyncWithPeers() error {
	if s.isSyncing {
		return errors.New("sync already in progress")
	}

	s.isSyncing = true
	defer func() { s.isSyncing = false }()

	peers := s.p2pServer.GetPeers()
	if len(peers) == 0 {
		return errors.New("no peers to sync with")
	}

	s.logger.Info("Starting blockchain sync...")

	// Get current height
	currentHeight := s.chain.GetHeight()

	// Query all peers for their heights
	peerHeights := make(map[string]uint64)
	for _, peer := range peers {
		height, err := s.getPeerHeight(peer)
		if err != nil {
			s.logger.Warnf("Failed to get height from peer %s: %v", peer.ID, err)
			continue
		}
		peerHeights[peer.ID] = height
	}

	// Find the best peer (highest height)
	var bestPeer *Peer
	var maxHeight uint64
	for _, peer := range peers {
		if height, ok := peerHeights[peer.ID]; ok && height > maxHeight {
			maxHeight = height
			bestPeer = peer
		}
	}

	if bestPeer == nil {
		return errors.New("no valid peers found")
	}

	if maxHeight <= currentHeight {
		s.logger.Info("Already in sync")
		return nil
	}

	s.logger.Infof("Syncing from peer %s (height %d -> %d)", bestPeer.ID, currentHeight, maxHeight)

	// Request blocks in batches
	batchSize := uint64(100)
	for height := currentHeight + 1; height <= maxHeight; height += batchSize {
		toHeight := height + batchSize - 1
		if toHeight > maxHeight {
			toHeight = maxHeight
		}

		blocks, err := s.requestBlocks(bestPeer, height, toHeight)
		if err != nil {
			return fmt.Errorf("failed to request blocks: %w", err)
		}

		// Validate and add blocks
		for _, block := range blocks {
			if err := s.chain.AddBlock(block); err != nil {
				return fmt.Errorf("failed to add block at height %d: %w", block.Header.Height, err)
			}
		}

		s.logger.Infof("Synced blocks %d to %d", height, toHeight)
	}

	s.logger.Info("Blockchain sync completed")
	return nil
}

// getPeerHeight requests the current height from a peer
func (s *Syncer) getPeerHeight(peer *Peer) (uint64, error) {
	msg := &Message{
		Type:    MsgTypeGetHeight,
		Payload: &GetHeightMessage{},
	}

	response, err := s.p2pServer.SendAndWaitForResponse(peer, msg, MsgTypeHeight, 10*time.Second)
	if err != nil {
		return 0, fmt.Errorf("failed to get peer height: %w", err)
	}

	// Parse response
	payloadBytes, err := json.Marshal(response.Payload)
	if err != nil {
		return 0, err
	}

	var heightMsg HeightMessage
	if err := json.Unmarshal(payloadBytes, &heightMsg); err != nil {
		return 0, err
	}

	return heightMsg.Height, nil
}

// requestBlocks requests blocks from a peer
func (s *Syncer) requestBlocks(peer *Peer, fromHeight, toHeight uint64) ([]*blockchain.Block, error) {
	msg := &Message{
		Type: MsgTypeGetBlocks,
		Payload: &GetBlocksMessage{
			FromHeight: fromHeight,
			ToHeight:   toHeight,
		},
	}

	response, err := s.p2pServer.SendAndWaitForResponse(peer, msg, MsgTypeBlocks, 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to request blocks: %w", err)
	}

	// Parse response
	payloadBytes, err := json.Marshal(response.Payload)
	if err != nil {
		return nil, err
	}

	var blocksMsg BlocksMessage
	if err := json.Unmarshal(payloadBytes, &blocksMsg); err != nil {
		return nil, err
	}

	return blocksMsg.Blocks, nil
}

// StartAutoSync starts automatic synchronization in the background
func (s *Syncer) StartAutoSync() {
	go func() {
		ticker := time.NewTicker(s.syncPeriod)
		defer ticker.Stop()

		for range ticker.C {
			if err := s.SyncWithPeers(); err != nil {
				s.logger.Warnf("Auto-sync failed: %v", err)
			}
		}
	}()
}

// TriggerSync triggers a sync if not already in progress
func (s *Syncer) TriggerSync() {
	if s.isSyncing {
		s.logger.Debug("Sync already in progress, skipping trigger")
		return
	}

	go func() {
		if err := s.SyncWithPeers(); err != nil {
			s.logger.Warnf("Triggered sync failed: %v", err)
		}
	}()
}

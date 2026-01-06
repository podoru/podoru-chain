package node

// NodeType defines the type of node
type NodeType string

const (
	// NodeTypeFull is a full node that validates all blocks
	NodeTypeFull NodeType = "full"

	// NodeTypeProducer is a block producer node (authority)
	NodeTypeProducer NodeType = "producer"
)

// String returns the string representation of node type
func (nt NodeType) String() string {
	return string(nt)
}

// IsValid checks if the node type is valid
func (nt NodeType) IsValid() bool {
	return nt == NodeTypeFull || nt == NodeTypeProducer
}

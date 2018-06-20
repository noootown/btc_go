package main

import "crypto/sha256"

type MerkleTree struct {
  rootNode *MerkleNode
}

type MerkleNode struct {
  left  *MerkleNode
  right *MerkleNode
  data  []byte
}

func NewMerkleTree(data [][]byte) *MerkleTree{
  var nodes []MerkleNode

  if len(data) % 2 != 0 {
    data = append(data, data[len(data)-1])
  }
  for _, datum := range data {
    nodes = append(nodes, *NewMerkleNode(nil, nil, datum))
  }

  for i := 0; i < len(data)/2; i++ {
    var newLevel []MerkleNode
    for j := 0; j < len(nodes); j += 2 {
      newLevel = append(newLevel, *NewMerkleNode(&nodes[j], &nodes[j+1], nil))
    }
    nodes = newLevel
  }
  return &MerkleTree{&nodes[0]}
}

// http://bit.ly/2tkCv9C
// slice of unaddressable value
func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode{
  var hash [32]byte
  if left == nil && right == nil {
    hash = sha256.Sum256(data)
  } else {
    hash = sha256.Sum256(append(left.data, right.data...))
  }
  return &MerkleNode{left, right, hash[:]}
}

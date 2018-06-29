package main

import (
  "time"
)

type Block struct {
  Timestamp int64
  Transactions []*Transaction
  PrevBlockHash []byte
  Hash []byte
  Nonce int
  Height int
}

func NewBlock(transactions []*Transaction, prevBlockHash []byte, height int) *Block {
  block := &Block{
    time.Now().Unix(),
    transactions,
    prevBlockHash,
    []byte(""), // ???
    0,
    height,
  }
  pow := NewProofOfWork(block)
  nonce, hash := pow.Run()

  block.Nonce = nonce
  block.Hash = hash

  return block
}

func NewGenesisBlock(coinbase *Transaction) *Block {
  return NewBlock([]*Transaction{coinbase}, []byte{}, 0)
}

func (block *Block) HashTransactions() []byte {
  var transactions [][]byte
  for _, tx := range block.Transactions {
    transactions = append(transactions, tx.Serialize())
  }

  return NewMerkleTree(transactions).rootNode.data
}

// Serialize serializes the block
func (b *Block) Serialize() []byte {
	return gobEncode(b)
}

func DeserializeBlock(data []byte) Block {
  var b Block
  gobDecode(data, &b)
  return b
}

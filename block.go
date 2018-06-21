package main

import (
  "time"
  "bytes"
  "encoding/gob"
  "log"
)

type Block struct {
  Timestamp int64
  Transactions []*Transaction
  PrevBlockHash []byte
  Hash []byte
  Nonce int
  // Height int
}

func NewBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
  block := &Block{
    time.Now().Unix(),
    transactions,
    prevBlockHash,
    []byte(""), // ???
    0,
  }
  pow := NewProofOfWork(block)
  nonce, hash := pow.Run()

  block.Nonce = nonce
  block.Hash = hash

  return block
}

func NewGenesisBlock(coinbase *Transaction) *Block {
  return NewBlock([]*Transaction{coinbase}, []byte{})
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
  var result bytes.Buffer
  err := gob.NewEncoder(&result).Encode(b)
  if err != nil {
    log.Panic(err)
  }
  return result.Bytes()
}

func DeserializeBlock(d []byte) *Block {
  var block Block
  err := gob.NewDecoder(bytes.NewReader(d)).Decode(&block)
  if err != nil {
    log.Panic(err)
  }
  return &block
}

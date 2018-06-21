package main

import (
  "github.com/boltdb/bolt"
  "log"
)

type BlockchainIterator struct {
  blockchain *Blockchain
  currentHash []byte
}

func NewBlockchainIterator(bc *Blockchain) *BlockchainIterator {
  b := BlockchainIterator{bc, nil}
  b.currentHash = bc.tip
  return &b
}

func (bci *BlockchainIterator) Next() *Block {
  var block *Block

  err := bci.blockchain.db.View(func(tx *bolt.Tx) error {
    block = DeserializeBlock(tx.Bucket([]byte(blocksBucket)).Get(bci.currentHash))
    return nil
  })
  if err != nil {
    log.Panic(err)
  }
  bci.currentHash = block.PrevBlockHash
  return block
}
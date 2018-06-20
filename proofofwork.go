package main

import (
  "math"
  "math/big"
  "bytes"
  "crypto/sha256"
)

const maxNonce = math.MaxInt64
const targetBits = 16

type ProofOfWork struct {
  block *Block
  target *big.Int
}

func NewProofOfWork(block *Block) *ProofOfWork{
  target := big.NewInt(1)
  target.Lsh(target, uint(256-targetBits))

  pow := &ProofOfWork{block, target}
  return pow
}

func (pow *ProofOfWork) PrepareData(nonce int) []byte {
  return bytes.Join(
    [][]byte{
      pow.block.PrevBlockHash,
      pow.block.HashTransactions(),
      IntToHex(pow.block.Timestamp),
      IntToHex(int64(targetBits)), // ????
      IntToHex(int64(nonce)),
    },
    []byte(""),
  )
}

func (pow *ProofOfWork) Run() (int, []byte){
  var hashInt big.Int
  var hash [32]byte
  nonce := 0

  for nonce < maxNonce {
    data := pow.PrepareData(nonce)
    hash = sha256.Sum256(data)
    hashInt.SetBytes(hash[:])
    if hashInt.Cmp(pow.target) == -1 {
      break
    }
    nonce++
  }
  return nonce, hash[:]
}

func (pow *ProofOfWork) Validate() bool {
  var hashInt big.Int
  hash := sha256.Sum256(pow.PrepareData(pow.block.Nonce))
  hashInt.SetBytes(hash[:])
  return hashInt.Cmp(pow.target) == -1
}
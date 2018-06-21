package main

import (
  "encoding/gob"
  "bytes"
  "log"
)

type TXOutput struct {
  Value int
  PubKeyHash []byte
}

func NewTXOutput(value int, address string) *TXOutput{
  pubKeyHash := Base58Decode([]byte(address))
  // 1 byte version + hash + 4 byte check code
  return &TXOutput{value,  pubKeyHash[1 : len(pubKeyHash)-4]}
}

func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool{
  return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

type TXOutputs struct {
  Outputs []TXOutput
}

func (out *TXOutputs) Serialize() []byte {
  var result bytes.Buffer
  err := gob.NewEncoder(&result).Encode(out)
  if err != nil {
    log.Panic(err)
  }
  return result.Bytes()
}

func DeserializeOutputs (data []byte) TXOutputs {
  var outputs TXOutputs
  err := gob.NewDecoder(bytes.NewReader(data)).Decode(&outputs)
  if err != nil {
    log.Panic(err)
  }
  return outputs
}

package main

import (
  "bytes"
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
  return gobEncode(out)
}

func DeserializeOutputs(data []byte) TXOutputs {
  var outputs TXOutputs
  gobDecode(data, &outputs)
  return outputs
}

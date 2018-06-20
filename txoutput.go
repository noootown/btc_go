package main

type TXOutput struct {
  Value int
  PubKeyHash []byte
}

func NewTXOutput(value int, address string) *TXOutput{
  pubKeyHash := Base58Decode([]byte(address))
  // 1 byte version + hash + 4 byte check code
  return &TXOutput{value,  pubKeyHash[1 : len(pubKeyHash)-4]}
}
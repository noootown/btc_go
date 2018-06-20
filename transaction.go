package main

import (
  "bytes"
  "encoding/gob"
  "log"
)

const subsidy = 10

type Transaction struct {
  ID [] byte
  Vin []TXInput
  Vout []TXOutput
}

func NewCoinbaseTX(address, data string) *Transaction{
  // todo
  // if data == "" {
  //
  // }

  txin := TXInput{[]byte{}, -1, nil, []byte(data)}
  txout := NewTXOutput(subsidy, address)
  return &Transaction{nil, []TXInput{txin}, []TXOutput{*txout}}
}

func (tx Transaction) IsCoinBase() bool {
  return len(tx.Vin) == 1 && len(tx.Vin[0].TXid) == 0 && tx.Vin[0].Vout == -1
}

func (tx Transaction) Serialize() []byte{
  var result bytes.Buffer
  err := gob.NewEncoder(&result).Encode(tx)
  if err != nil {
    log.Panic(err)
  }
  return result.Bytes()
}

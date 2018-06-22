package main

import (
  "bytes"
  "encoding/gob"
  "log"
  "crypto/sha256"
  "fmt"
  "strings"
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
  tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}}
  tx.ID = tx.Hash()
  return &tx
}

func (tx *Transaction) Hash() []byte {
  var hash [32]byte
  txCopy := *tx
  txCopy.ID = []byte{}
  hash = sha256.Sum256(txCopy.Serialize())
  return hash[:]
}

func (tx Transaction) IsCoinbase() bool {
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

func (tx Transaction) String() string {
  var lines []string

  lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))

  for i, input := range tx.Vin {
    lines = append(lines, fmt.Sprintf("     Input %d:", i))
    lines = append(lines, fmt.Sprintf("       TXID:      %x", input.TXid))
    lines = append(lines, fmt.Sprintf("       Out:       %d", input.Vout))
    lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
    lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
  }

  for i, output := range tx.Vout {
    lines = append(lines, fmt.Sprintf("     Output %d:", i))
    lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
    lines = append(lines, fmt.Sprintf("       Script: %x", output.PubKeyHash))
  }

  return strings.Join(lines, "\n")
}

package main

import (
  "fmt"
  "os"
  "io/ioutil"
  "log"
  "encoding/gob"
  "crypto/elliptic"
  "bytes"
)

const walletFile = "wallet_%s.dat"

type Wallets struct {
  Wallets map[string]*Wallet
}

func (ws *Wallets) CreateWallet() string {
  wallet := NewWallet()
  address := fmt.Sprintf("%s", wallet.GetAddress())
  ws.Wallets[address] = wallet
  return address
}

func (ws *Wallets) GetWallet(address string) Wallet {
  return *ws.Wallets[address]
}

func (ws *Wallets) GetAddresses() []string {
  var addresses []string

  for address := range ws.Wallets {
    addresses = append(addresses, address)
  }

  return addresses
}

func NewWallets(nodeID string) (*Wallets, error) {
  wallets := Wallets{}
  wallets.Wallets = make(map[string]*Wallet)

  walletFile := fmt.Sprintf(walletFile, nodeID)
  if _, err := os.Stat(walletFile); os.IsNotExist(err) {
    return &wallets, err
  }

  fileContent, err := ioutil.ReadFile(walletFile)
  if err != nil {
    log.Panic(err)
  }

  gob.Register(elliptic.P256())
  err = gob.NewDecoder(bytes.NewReader(fileContent)).Decode(&wallets)
  if err != nil {
    log.Panic(err)
  }

  return &wallets, nil
}

func (ws *Wallets) SaveToFile(nodeID string) { // ?????
  var content bytes.Buffer
  walletFile := fmt.Sprintf(walletFile, nodeID)

  gob.Register(elliptic.P256())

  err := gob.NewEncoder(&content).Encode(ws)
  if err != nil {
    log.Panic(err)
  }

  err = ioutil.WriteFile(walletFile, content.Bytes(), 0644)
  if err != nil {
    log.Panic(err)
  }

}

package main

import (
  "github.com/boltdb/bolt"
  "fmt"
  "os"
  "log"
)

const dbFile = "blockchain_%s.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

type Blockchain struct {
  tip []byte
  db *bolt.DB // ?????
}

func createBlockchainDB(address, nodeID string) *Blockchain{
  dbFile := fmt.Sprintf(dbFile, nodeID)
  if IsDBExist(dbFile) {
    fmt.Println("Blockchain already exists.")
    os.Exit(1)
  }
  var tip []byte

  genesis := NewGenesisBlock(NewCoinbaseTX(address, genesisCoinbaseData))

  db, err := bolt.Open(dbFile, 0600, nil)
  if err != nil {
    log.Panic(err)
  }

  err = db.Update(func(tx *bolt.Tx) error {
    b, err := tx.CreateBucket([]byte(blocksBucket))
    if err != nil {
      log.Panic(err)
    }

    err = b.Put(genesis.Hash, genesis.Serialize())
    if err != nil {
      log.Panic(err)
    }

    err = b.Put([]byte("l"), genesis.Hash)
    if err != nil {
      log.Panic(err)
    }
    tip = genesis.Hash

    return nil
  })
  if err != nil {
    log.Panic(err)
  }
  bc := Blockchain{tip, db}
  return &bc
}

// func NewBlockchain(nodeID string) *Blockchain {
//   dbFile := fmt.Sprintf(dbFile, nodeID)
//   if !IsDBExist(dbFile) {
//     fmt.Println("No existing blockchain found. Create one first.")
//     os.Exit(1)
//   }
//   var tip []byte
//   db, err := bolt.Open(dbFile, 0600, nil)
//   if err != nil {
//     log.Panic(err)
//   }
//   err = db.Update(func(tx *bolt.Tx) error { // ???bolt.Tx
//     tip = tx.Bucket([]byte(blocksBucket)).Get([]byte("l"))
//     return nil
//   })
//   if err != nil {
//     log.Panic(err)
//   }
//   return &Blockchain{tip, db}
// }

func IsDBExist(dbfile string) bool{
  _, err := os.Stat(dbfile)
  return !os.IsNotExist(err)
}

package main

import (
  "github.com/boltdb/bolt"
  "log"
  "encoding/hex"
)

const utxoBucket = "chainstate"

type UTXOSet struct {
  Blockchain *Blockchain
}

func (u UTXOSet) FindUTXO(pubKeyHash []byte) []TXOutput {
  var UTXOs []TXOutput

  err := u.Blockchain.db.View(func(tx *bolt.Tx) error {
    c := tx.Bucket([]byte(utxoBucket)).Cursor()

    for k, v := c.First(); k != nil; k, v = c.Next() {
      outs := DeserializeOutputs(v)

      for _, out := range outs.Outputs {
        if out.IsLockedWithKey(pubKeyHash) {
          UTXOs = append(UTXOs, out)
        }
      }
    }
    return nil
  })
  if err != nil {
    log.Panic(err)
  }
  return UTXOs
}

func (u UTXOSet) Reindex() { // ????? pointer or not
  db := u.Blockchain.db
  bucketName := []byte(utxoBucket)

  err := db.Update(func(tx *bolt.Tx) error {
    err := tx.DeleteBucket(bucketName)
    if err != nil && err != bolt.ErrBucketNotFound {
      log.Panic(err)
    }

    _, err = tx.CreateBucket(bucketName)
    if err != nil {
      log.Panic(err)
    }

    return nil
  })
  if err != nil {
    log.Panic(err)
  }

  UTXO := u.Blockchain.FindUTXO()

  err = db.Update(func(tx *bolt.Tx) error {
    for txID, outs := range UTXO {
      key, err := hex.DecodeString(txID)
      if err != nil {
        log.Panic(err)
      }
      err = tx.Bucket(bucketName).Put(key, outs.Serialize())
      if err != nil {
        log.Panic(err)
      }
    }
    return nil
  })
}

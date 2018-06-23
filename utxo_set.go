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

func (u *UTXOSet) FindUTXO(pubKeyHash []byte) []TXOutput {
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

func (u *UTXOSet) Reindex() {
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
func (u* UTXOSet) CountTransactions() int {
  counter := 0
  err := u.Blockchain.db.View(func(tx *bolt.Tx) error {
    c := tx.Bucket([]byte(utxoBucket)).Cursor()
    for k, _ := c.First(); k != nil; k, _ = c.Next() {
      counter++
    }
    return nil
  })
  if err != nil {
    log.Panic(err)
  }

  return counter
}

func (u *UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int){
  unspentOutputs := make(map[string][]int)
  accumulated := 0

  err := u.Blockchain.db.View(func(tx *bolt.Tx) error {
    c := tx.Bucket([]byte(utxoBucket)).Cursor()
    Work:
      for k, v := c.First(); k != nil; k, v = c.Next() {
        txID := hex.EncodeToString(k)
        outs := DeserializeOutputs(v)
        for outIdx, out := range outs.Outputs {
          if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
            accumulated += out.Value
            unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
            if accumulated >= amount {
              break Work
            }
          }
        }
      }
    return nil
  })

  if err != nil {
    log.Panic(err)
  }
  return accumulated, unspentOutputs
}

func (u *UTXOSet) Update(block *Block) {
  err := u.Blockchain.db.Update(func (tx *bolt.Tx) error {
    b := tx.Bucket([]byte(utxoBucket))

    for _, tx := range block.Transactions {
      if !tx.IsCoinbase() {
        for _, vin := range tx.Vin {
          updatedOuts := TXOutputs{}
          outsBytes := b.Get(vin.TXid)
          outs := DeserializeOutputs(outsBytes)

          for outIdx, out := range outs.Outputs {
            if outIdx != vin.Vout {
              updatedOuts.Outputs = append(updatedOuts.Outputs, out)
            }
          }
          if len(updatedOuts.Outputs) == 0 {
            err := b.Delete(vin.TXid)
            if err != nil {
              log.Panic(err)
            }
          } else {
            err := b.Put(vin.TXid, updatedOuts.Serialize())
            if err != nil {
              log.Panic(err)
            }
          }
        }
      }
      newOutputs := TXOutputs{}
      for _, out := range tx.Vout {
        newOutputs.Outputs = append(newOutputs.Outputs, out)
      }
      err := b.Put(tx.ID, newOutputs.Serialize())
      if err != nil {
        log.Panic(err)
      }
    }
    return nil
  })
  if err != nil {
    log.Panic(err)
  }
}

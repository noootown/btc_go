package main

import (
  "github.com/boltdb/bolt"
  "fmt"
  "os"
  "log"
  "encoding/hex"
  "crypto/ecdsa"
  "bytes"
  "errors"
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
  if IsFileExist(dbFile) {
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
  return &Blockchain{tip, db}
}

func NewBlockchain(nodeID string) *Blockchain {
  dbFile := fmt.Sprintf(dbFile, nodeID)
  if !IsFileExist(dbFile) {
    fmt.Println("No existing blockchain found. Create one first.")
    os.Exit(1)
  }

  var tip []byte
  db, err := bolt.Open(dbFile, 0600, nil)
  if err != nil {
    log.Panic(err)
  }

  err = db.Update(func(tx *bolt.Tx) error {
    tip = tx.Bucket([]byte(blocksBucket)).Get([]byte("l"))
    return nil
  })
  if err != nil {
    log.Panic(err)
  }

  return &Blockchain{tip, db}
}

func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
  bci := bc.Iterator()

  for {
    block := bci.Next()
    for _, tx := range block.Transactions {
      if bytes.Compare(tx.ID, ID) == 0 {
        return *tx, nil
      }
    }
    if len(block.PrevBlockHash) == 0 {
      break
    }
  }
  return Transaction{}, errors.New("Transaction is not found")
}

func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
  prevTXs := make(map[string]Transaction)

  for _, vin := range tx.Vin {
    prevTX, err := bc.FindTransaction(vin.TXid)
    if err != nil {
      log.Panic(err)
    }
    prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
  }

  tx.Sign(privKey, prevTXs)
}

func (bc *Blockchain) FindUTXO() map[string]TXOutputs {
  UTXO := make(map[string]TXOutputs)
  spentTXOs := make(map[string][]int)
  bci := bc.Iterator()
  for {
    block := bci.Next()
    for _, tx := range block.Transactions {
      txID := hex.EncodeToString(tx.ID)

    Outputs:
      for outIdx, out := range tx.Vout {
        if spentTXOs[txID] != nil {
          for _, spentOutIdx := range spentTXOs[txID] {
            if spentOutIdx == outIdx {
              continue Outputs
            }
          }
        }
        outs := UTXO[txID]
        outs.Outputs = append(outs.Outputs, out)
        UTXO[txID] = outs
      }

      if !tx.IsCoinbase() {
        for _, in := range tx.Vin {
          inTxID := hex.EncodeToString(in.TXid)
          spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
        }
      }
    }
    if len(block.PrevBlockHash) == 0 {
      break
    }
  }
  return UTXO
}

type BlockchainIterator struct {
  blockchain *Blockchain
  currentHash []byte
}

func (bci *BlockchainIterator) Next() *Block {
  var block *Block

  err := bci.blockchain.db.View(func(tx *bolt.Tx) error {
    block = DeserializeBlock(tx.Bucket([]byte(blocksBucket)).Get(bci.currentHash))
    return nil
  })
  if err != nil {
    log.Panic(err)
  }
  bci.currentHash = block.PrevBlockHash
  return block
}

func (bc *Blockchain) Iterator() *BlockchainIterator {
  return &BlockchainIterator{bc, bc.tip}
}

func (bc *Blockchain) MineBlock(txs []*Transaction) *Block{
  var tip []byte
  // var lastHeight int

  err := bc.db.View(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte(blocksBucket))
    tip = b.Get([]byte("l"))

    return nil
  })
  if err != nil {
    log.Panic(err)
  }
  newBlock := NewBlock(txs, tip)

  err = bc.db.Update(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte(blocksBucket))
    err := b.Put(newBlock.Hash, newBlock.Serialize())
    if err != nil {
      log.Panic(err)
    }

    err = b.Put([]byte("l"), newBlock.Hash)
    if err != nil {
      log.Panic(err)
    }

    bc.tip = newBlock.Hash

    return nil
  })
  if err != nil {
    log.Panic(err)
  }
  return newBlock
}

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

  for !bci.IsDone {
    block := bci.Next()
    for _, tx := range block.Transactions {
      if bytes.Compare(tx.ID, ID) == 0 {
        return *tx, nil
      }
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
  for !bci.IsDone {
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
  }
  return UTXO
}

type BlockchainIterator struct {
  blockchain *Blockchain
  currentHash []byte
  IsDone bool
}

func (bci *BlockchainIterator) Next() *Block {
  var block Block

  err := bci.blockchain.db.View(func(tx *bolt.Tx) error {
    block = DeserializeBlock(tx.Bucket([]byte(blocksBucket)).Get(bci.currentHash))
    return nil
  })
  if err != nil {
    log.Panic(err)
  }
  bci.currentHash = block.PrevBlockHash
  bci.IsDone = len(block.PrevBlockHash) == 0
  return &block
}

func (bc *Blockchain) Iterator() *BlockchainIterator {
  return &BlockchainIterator{bc, bc.tip, false}
}

func (bc *Blockchain) MineBlock(txs []*Transaction) *Block{
  var tip []byte
  var lastHeight int

  err := bc.db.View(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte(blocksBucket))
    tip = b.Get([]byte("l"))
    lastHeight = DeserializeBlock(b.Get(tip)).Height

    return nil
  })
  if err != nil {
    log.Panic(err)
  }
  newBlock := NewBlock(txs, tip, lastHeight+1)

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

func (bc *Blockchain) GetBestHeight() int{
  var lastBlock Block

  err := bc.db.View(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte(blocksBucket))
    lastBlock = DeserializeBlock(b.Get(b.Get([]byte("l"))))
    return nil
  })
  if err != nil {
    log.Panic(err)
  }
  return lastBlock.Height
}

func (bc *Blockchain) GetBlock(blockHash []byte) (Block, error) {
  var block Block
  err := bc.db.View(func(tx *bolt.Tx) error {
    blockData := tx.Bucket([]byte(blocksBucket)).Get(blockHash)
    if blockData == nil {
      return errors.New("Block is not found.")
    }
    block = DeserializeBlock(blockData)
    return nil
  })
  if err != nil {
    return block, err
  }
  return block, nil
}

func (bc *Blockchain) GetBlockHashes() [][]byte {
  var blocks [][]byte
  bci := bc.Iterator()

  for !bci.IsDone {
    block := bci.Next()
    blocks = append(blocks, block.Hash)
  }
  return blocks
}

func (bc *Blockchain) AddBlock(block *Block) {
  err := bc.db.Update(func(tx *bolt.Tx) error {
    b := tx.Bucket([]byte(blocksBucket))
    blockInDb := b.Get(block.Hash)

    if blockInDb != nil {
      return nil
    }

    err := b.Put(block.Hash, block.Serialize())
    if err != nil {
      log.Panic(err)
    }

    lastBlock := DeserializeBlock(b.Get(b.Get([]byte("l"))))

    if block.Height > lastBlock.Height {
      err = b.Put([]byte("l"), block.Hash)
      if err != nil {
        log.Panic(err)
      }
      bc.tip = block.Hash
    }
    return nil
  })
  if err != nil {
    log.Panic(err)
  }
}

func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
  if tx.IsCoinbase() {
    return true
  }
  prevTXs := make(map[string]Transaction)
  for _, vin := range tx.Vin {
    prevTX, err := bc.FindTransaction(vin.TXid)
    if err != nil {
      log.Panic(err)
    }
    prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
  }

  return tx.Verify(prevTXs)

}

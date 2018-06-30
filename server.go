package main

import (
  "fmt"
  "net"
  "log"
  "io"
  "bytes"
  "io/ioutil"
  "encoding/hex"
)

const protocol = "tcp"
const nodeVersion = 1
const commandLength = 12

var nodeAddress string
var miningAddress string
var knownNodes = []string{"localhost:3000"}
var blocksInTransit = [][]byte{}
var mempool = make(map[string]Transaction)

type block struct {
  AddrFrom string
  Block    []byte
}

type verzion struct {
  Version    int
  BestHeight int
  AddrFrom   string
}

type getblocks struct {
  AddrFrom string
}

type inv struct {
  AddrFrom string
  Type     string
  Items    [][]byte
}

type getdata struct {
  AddrFrom string
  Type     string
  ID    []byte
}

type tx struct {
  AddFrom     string
  Transaction []byte
}

func commandToBytes(command string) []byte {
  var bytes [commandLength]byte
  for i, c := range command {
    bytes[i] = byte(c)
  }
  return bytes[:]
}

func bytesToCommand(bytes []byte) string {
  var command []byte
  for _, b := range bytes {
    if b != 0x0 {
      command = append(command, b)
    }
  }
  return fmt.Sprintf("%s", command)
}

func StartServer(nodeID, minerAddress string) {
  nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
  miningAddress = minerAddress
  ln, err := net.Listen(protocol, nodeAddress)
  if err != nil {
    log.Panic(err)
  }
  defer ln.Close()

  bc := NewBlockchain(nodeID)
  if nodeAddress != knownNodes[0] {
    sendVersion(knownNodes[0], bc)
  }
  for {
    conn, err := ln.Accept()
    if err != nil {
      log.Panic(err)
    }
    go handleConnection(conn, bc)
  }
}
func sendBlock(address string, b *Block) {
  data := block{nodeAddress, b.Serialize()}
  payload := gobEncode(data)
  request := append(commandToBytes("block"), payload...)
  sendData(address, request)
}

func sendInv(address, kind string, items [][]byte) {
  inventory := inv{nodeAddress, kind, items}
  payload := gobEncode(inventory)
  request := append(commandToBytes("inv"), payload...)
  sendData(address, request)
}

func sendGetBlocks(address string) {
  payload := gobEncode(getblocks{nodeAddress})
  request := append(commandToBytes("getblocks"), payload...)
  sendData(address, request)
}

func sendTx(address string, tnx *Transaction) {
  payload := gobEncode(tx{nodeAddress, tnx.Serialize()})
  request := append(commandToBytes("tx"), payload...)
  sendData(address, request)
}

func sendGetData(address, kind string, hash []byte) {
  payload := gobEncode(getdata{nodeAddress, kind, hash})
  request := append(commandToBytes("getdata"), payload...)
  sendData(address, request)
}

func sendVersion(address string, bc *Blockchain) {
  bestHeight := bc.GetBestHeight()
  payload := gobEncode(verzion{nodeVersion, bestHeight, nodeAddress})
  request := append(commandToBytes("version"), payload...)
  sendData(address, request)
}

func sendData(addr string, data []byte) {
  conn, err := net.Dial(protocol, addr)
  if err != nil {
    fmt.Printf("%s is not available\n", addr)
    var updatedNodes []string
    for _, node := range knownNodes {
      if node != addr {
        updatedNodes = append(updatedNodes, node)
      }
    }
    knownNodes = updatedNodes
    return
  }
  defer conn.Close()

  _, err = io.Copy(conn, bytes.NewReader(data))
  if err != nil {
    log.Panic(err)
  }
}

func handleConnection(conn net.Conn, bc *Blockchain) {
  request, err := ioutil.ReadAll(conn)
  if err != nil {
    log.Panic(err)
  }
  command := bytesToCommand(request[:commandLength])
  fmt.Printf("Received %s command\n", command)

  switch command {
  //case "addr":
  //  handleAddr(request)
  case "block":
    handleBlock(request, bc)
  case "inv":
    handleInv(request, bc)
  case "getblocks":
    handleGetBlocks(request, bc)
  case "getdata":
    handleGetData(request, bc)
  case "tx":
     handleTx(request, bc)
  case "version":
    handleVersion(request, bc)
  default:
    fmt.Println("Unknown command!")
  }
  conn.Close()
}

func handleBlock(request []byte, bc *Blockchain) {
  var payload block
  gobDecode(request[commandLength:], &payload)

  block := DeserializeBlock(payload.Block)

  fmt.Println("Recevied a new block!")
  bc.AddBlock(&block)

  fmt.Printf("Added block %x\n", block.Hash)

  if len(blocksInTransit) > 0 {
    blockHash := blocksInTransit[0]
    sendGetData(payload.AddrFrom, "block", blockHash)
    blocksInTransit = blocksInTransit[1:]
  } else {
    UTXOSet := UTXOSet{bc}
    UTXOSet.Reindex()
  }
}

func handleInv(request []byte, bc *Blockchain) {
  var payload inv
  gobDecode(request[commandLength:], &payload)

  fmt.Printf("Recevied inventory with %d %s\n", len(payload.Items), payload.Type)
  if payload.Type == "block" {
    blocksInTransit = payload.Items

    blockHash := payload.Items[0]
    sendGetData(payload.AddrFrom, "block", blockHash)

    newInTransit := [][]byte{}
    for _, b := range blocksInTransit {
      if bytes.Compare(b, blockHash) != 0 {
        newInTransit = append(newInTransit, b)
      }
    }
    blocksInTransit = newInTransit
  } else if payload.Type == "tx" {
    txID := payload.Items[0]
    if mempool[hex.EncodeToString(txID)].ID == nil {
      sendGetData(payload.AddrFrom, "tx", txID)
    }
  }
}

func handleGetBlocks(request []byte, bc *Blockchain) {
  var payload getblocks
  gobDecode(request[commandLength:], &payload)
  blocks := bc.GetBlockHashes()
  sendInv(payload.AddrFrom, "block", blocks)
}

func handleGetData(request []byte, bc *Blockchain) {
  var payload getdata
  gobDecode(request[commandLength:], &payload)

  if payload.Type == "block" {
    block, err := bc.GetBlock([]byte(payload.ID))
    if err != nil {
      return
    }
    sendBlock(payload.AddrFrom, &block)
  } else if payload.Type == "tx" {
    tx := mempool[hex.EncodeToString(payload.ID)]
    sendTx(payload.AddrFrom, &tx)
  }
}

func handleTx(request []byte, bc *Blockchain) {
  var payload tx
  gobDecode(request[commandLength:], &payload)

  tx := DeserializeTransaction(payload.Transaction)
  mempool[hex.EncodeToString(tx.ID)] = tx

  if nodeAddress == knownNodes[0] {
    for _, node := range knownNodes {
      if node != nodeAddress && node != payload.AddFrom {
        sendInv(node, "tx", [][]byte{tx.ID})
      }
    }
  } else {
    if len(mempool) >= 2 && len(miningAddress) > 0 {
      for len(mempool) > 0 {
        var txs []*Transaction
        for id := range mempool {
          tx := mempool[id]
          if bc.VerifyTransaction(&tx) {
            txs = append(txs, &tx)
          }
        }

        if len(txs) == 0 {
          fmt.Println("All transactions are invalid! Waiting for new ones...")
          return
        }

        txs = append(txs, NewCoinbaseTX(miningAddress, ""))

        newBlock := bc.MineBlock(txs)
        UTXOSet := UTXOSet{bc}
        UTXOSet.Reindex()

        fmt.Println("New block is mined!")

        for _, tx := range txs {
          delete(mempool, hex.EncodeToString(tx.ID))
        }

        for _, node := range knownNodes {
          if node != nodeAddress {
            sendInv(node, "block", [][]byte{newBlock.Hash})
          }
        }
      }
    }
  }
}

func handleVersion(request []byte, bc *Blockchain) {
  var payload verzion
  gobDecode(request[commandLength:], &payload)

  myBestHeight := bc.GetBestHeight()
  foreignerBestHeight := payload.BestHeight

  if myBestHeight < foreignerBestHeight {
    sendGetBlocks(payload.AddrFrom)
  } else if myBestHeight > foreignerBestHeight {
    sendVersion(payload.AddrFrom, bc)
  }

  // sendAddr(payload.AddrFrom)
  if !nodeIsKnown(payload.AddrFrom) {
    knownNodes = append(knownNodes, payload.AddrFrom)
  }
}

func nodeIsKnown(addr string) bool {
  for _, node := range knownNodes {
    if node == addr {
      return true
    }
  }
  return false
}

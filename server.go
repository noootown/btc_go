package main

import (
	"fmt"
	"net"
	"log"
  "io"
  "bytes"
  "io/ioutil"
)

const protocol = "tcp"
const nodeVersion = 1
const commandLength = 12
var nodeAddress string
var miningAddress string
var knownNodes = []string{"localhost:3000"}

type verzion struct {
	Version    int
	BestHeight int
	AddrFrom   string
}

type getblocks struct {
  AddrFrom string
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

func sendVersion(addr string, bc *Blockchain) {
	bestHeight := bc.GetBestHeight()
	payload := gobEncode(verzion{nodeVersion, bestHeight, nodeAddress})
	request := append(commandToBytes("version"), payload...)
	sendData(addr, request)
}

func sendGetBlocks(address string) {
  payload := gobEncode(getblocks{nodeAddress})
  request := append(commandToBytes("getblocks"), payload...)

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
  //case "block":
  //  handleBlock(request, bc)
  //case "inv":
  //  handleInv(request, bc)
  case "getblocks":
   handleGetBlocks(request, bc)
  //case "getdata":
  //  handleGetData(request, bc)
  //case "tx":
  //  handleTx(request, bc)
  case "version":
    handleVersion(request, bc)
  default:
    fmt.Println("Unknown command!")
  }
  conn.Close()
}
func handleGetBlocks(request []byte, bc *Blockchain) {
  var payload getblocks
  gobDecode(request[commandLength:], payload)
  //blocks := bc.GetBlockHashes()
  //sendInv(payload.AddrFrom, "block", blocks)
}

func handleVersion(request []byte, bc *Blockchain) {
  var payload verzion
  gobDecode(request[commandLength:], payload)

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

package main

import (
  "os"
  "fmt"
  "flag"
  "log"
  "strconv"
)

type CLI struct {}
func (cli *CLI) printUsage() {
  fmt.Println(`Usage:
    createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS
    createwallet - Generates a new key-pair and saves it into the wallet file
    getbalance -address ADDRESS - Get balance of ADDRESS
    listaddresses - Lists all addresses from the wallet file
    printchain - Print all the blocks of the blockchain
    reindexutxo - Rebuilds the UTXO set
    send -from FROM -to TO -amount AMOUNT -mine - Send AMOUNT of coins from FROM address to TO. Mine on the same node, when -mine is set.
    startnode -miner ADDRESS - Start a node with ID specified in NODE_ID env. var. -miner enables mining
    listbalance - Lists balance of addresses from the wallet file
  `)
}

func (cli *CLI) Run() {
  if len(os.Args) < 2 {
    cli.printUsage()
    os.Exit(1)
  }

  nodeID := os.Getenv("NODE_ID")
  if nodeID == "" {
    fmt.Printf("NODE_ID env. var is not set!")
    os.Exit(1)
  }

  cmdOption := []string{
    "createblockchain", "createwallet", "getbalance", "listaddresses",
    "printchain", "reindexutxo", "send", "startnode", "listbalance",
  }
  var cmd = make(map[string]*flag.FlagSet)
  for _, c := range cmdOption {
    cmd[c] = flag.NewFlagSet(c, flag.ExitOnError)
  }

  createBlockchainAddress := cmd["createblockchain"].String("address", "", "The address to send genesis block reward to")
  getBalanceAddress := cmd["getbalance"].String("address", "", "The address to get balance for")
  sendFrom := cmd["send"].String("from", "", "Source wallet address")
  sendTo := cmd["send"].String("to", "", "Destination wallet address")
  sendAmount := cmd["send"].Int("amount", 0, "Amount to send")
  sendMine := cmd["send"].Bool("mine", false, "Mine immediately on the same node")
  minerAddress := cmd["startnode"].String("miner", "", "Enable mining mode and send reward to ADDRESS")
  command := os.Args[1]
  if !stringInSlice(command, cmdOption) {
    cli.printUsage()
    os.Exit(1)
  }
  err := cmd[command].Parse(os.Args[2:])
  if err != nil {
    cli.printUsage()
    log.Panic(err)
  }

  if cmd["createblockchain"].Parsed() {
    address := *createBlockchainAddress
    if address == "" {
      cmd["createblockchain"].Usage()
      os.Exit(1)
    }
    if !IsAddressValid(address) {
      log.Panic("ERROR: Address is not valid")
    }
    bc := createBlockchainDB(address, nodeID)
    defer bc.db.Close()
    UTXOSet := UTXOSet{bc}
    UTXOSet.Reindex()

    fmt.Printf("Done with creating blockchain with genesis block at address %s!\n", address)
  } else if cmd["getbalance"].Parsed() {
    address := *getBalanceAddress
    if address == "" {
      cmd["getbalance"].Usage()
      os.Exit(1)
    }
    if !IsAddressValid(address) {
      log.Panic("ERROR: Address is not valid")
    }
    bc := NewBlockchain(nodeID)
    UTXOSet := UTXOSet{bc}
    defer bc.db.Close()

    balance := 0
    pubKeyHash := Base58Decode([]byte(address))
    pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
    UTXOs := UTXOSet.FindUTXO(pubKeyHash)

    for _, out := range UTXOs {
      balance += out.Value
    }
    fmt.Printf("Balance of '%s': %d\n", address, balance)
  } else if cmd["createwallet"].Parsed() {
    wallets, _ := NewWallets(nodeID)
    address := wallets.CreateWallet()
    wallets.SaveToFile(nodeID)
    fmt.Printf("Your new address: %s\n", address)
  } else if cmd["listaddresses"].Parsed() {
    wallets, err := NewWallets(nodeID)
    if err != nil {
      log.Panic(err)
    }
    addresses := wallets.GetAddresses()

    for _, address := range addresses {
      fmt.Println(address)
    }
  } else if cmd["printchain"].Parsed() {
    bc := NewBlockchain(nodeID)
    bci := bc.Iterator()
    defer bc.db.Close()
    for !bci.IsDone {
      block := bci.Next()
      fmt.Printf("============ Block %x ============\n", block.Hash)
      // fmt.Printf("Height: %d\n", block.Height)
      fmt.Printf("Prev. block: %x\n", block.PrevBlockHash)
      pow := NewProofOfWork(block)
      fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))
      for _, tx := range block.Transactions {
        fmt.Println(tx)
      }
      fmt.Printf("\n\n")
    }
  } else if cmd["reindexutxo"].Parsed() {
    bc := NewBlockchain(nodeID)
    UTXOSet := UTXOSet{bc}
    UTXOSet.Reindex()

    fmt.Printf("Done! There are %d transactions in the UTXO set.\n", UTXOSet.CountTransactions())
  } else if cmd["send"].Parsed() {
    from, to, amount, mineNow := *sendFrom, *sendTo, *sendAmount, *sendMine
    if from == "" || to == "" || amount <= 0 {
      cmd["send"].Usage()
      os.Exit(1)
    }
    if !IsAddressValid(from) {
      log.Panic("ERROR: Sender address is not valid")
    }
    if !IsAddressValid(to) {
      log.Panic("ERROR: Recipient address is not valid")
    }

    bc := NewBlockchain(nodeID)
    UTXOSet := UTXOSet{bc}
    defer bc.db.Close()

    wallets, err := NewWallets(nodeID)
    if err != nil {
      log.Panic(err)
    }
    wallet := wallets.GetWallet(from)
    tx := NewUTXOTransaction(&wallet, to, amount, &UTXOSet)
    if mineNow {
      txs := []*Transaction{NewCoinbaseTX(from, ""), tx}

      UTXOSet.Update(bc.MineBlock(txs))
    } else {
      sendTx(knownNodes[0], tx)
    }
  } else if cmd["startnode"].Parsed() {
    nodeID := os.Getenv("NODE_ID")
    if nodeID == "" {
      cmd["startnode"].Usage()
      os.Exit(1)
    }
    fmt.Printf("Starting node %s\n", nodeID)
    if len(*minerAddress) > 0 {
      if IsAddressValid(*minerAddress) {
        fmt.Println("Mining is on. Address to receive rewards: ", *minerAddress)
      } else {
        log.Panic("Wrong miner address!")
      }
    }
    StartServer(nodeID, *minerAddress)
  } else if cmd["listbalance"].Parsed() {
    wallets, err := NewWallets(nodeID)
    if err != nil {
      log.Panic(err)
    }
    addresses := wallets.GetAddresses()
    bc := NewBlockchain(nodeID)
    UTXOSet := UTXOSet{bc}
    defer bc.db.Close()

    for _, address := range addresses {
      balance := 0
      pubKeyHash := Base58Decode([]byte(address))
      pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
      UTXOs := UTXOSet.FindUTXO(pubKeyHash)

      for _, out := range UTXOs {
        balance += out.Value
      }
      fmt.Printf("%s: %d\n", address, balance)
    }
  }
}

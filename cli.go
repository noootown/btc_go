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
    startnode -miner ADDRESS - Start a node with ID specified in NODE_ID env. var. -miner enables mining`)
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
    "printchain", "reindexutxo", "send", "startnode",
  }
  var cmd []*flag.FlagSet
  for _, c := range cmdOption {
    cmd = append(cmd, flag.NewFlagSet(c, flag.ExitOnError))
  }
  createBlockchainCmd := cmd[0]
  createWalletCmd :=  cmd[1]
  getBalanceCmd := cmd[2]
  listAddressesCmd := cmd[3]
  printChainCmd := cmd[4]
  reindexUTXOCmd := cmd[5]
  sendCmd := cmd[6]
  // startNodeCmd := cmd[7]

  getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
  createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
  sendFrom := sendCmd.String("from", "", "Source wallet address")
  sendTo := sendCmd.String("to", "", "Destination wallet address")
  sendAmount := sendCmd.Int("amount", 0, "Amount to send")
  sendMine := sendCmd.Bool("mine", false, "Mine immediately on the same node")
  // minerAddress := startNodeCmd.String("miner", "", "Enable mining mode and send reward to ADDRESS")
  isDone := false
  for i, c := range cmdOption {
    if os.Args[1] == c {
      err := cmd[i].Parse(os.Args[2:])
      if err != nil {
        log.Panic(err)
      }
      isDone = true
      break
    }
  }
  if !isDone {
    cli.printUsage()
    os.Exit(1)
  }

  if createBlockchainCmd.Parsed() {
    address := *createBlockchainAddress
    if address == "" {
      createBlockchainCmd.Usage()
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
  }

  if getBalanceCmd.Parsed() {
    address := *getBalanceAddress
    if address == "" {
      getBalanceCmd.Usage()
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
  }

  if createWalletCmd.Parsed() {
    wallets, _ := NewWallets(nodeID)
    address := wallets.CreateWallet()
    wallets.SaveToFile(nodeID)
    fmt.Printf("Your new address: %s\n", address)
  }

  if listAddressesCmd.Parsed() {
    wallets, err := NewWallets(nodeID)
    if err != nil {
      log.Panic(err)
    }
    addresses := wallets.GetAddresses()

    for _, address := range addresses {
      fmt.Println(address)
    }
  }

  if printChainCmd.Parsed() {
    bc := NewBlockchain(nodeID)
    bci := bc.Iterator()
    defer bc.db.Close()
    for {
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
      if len(block.PrevBlockHash) == 0 {
        break
      }
    }
  }

  if reindexUTXOCmd.Parsed() {
    bc := NewBlockchain(nodeID)
    UTXOSet := UTXOSet{bc}
    UTXOSet.Reindex()

    fmt.Printf("Done! There are %d transactions in the UTXO set.\n", UTXOSet.CountTransactions())
  }

  if sendCmd.Parsed() {
    from, to, amount, mineNow := *sendFrom, *sendTo, *sendAmount, *sendMine
    if from == "" || to == "" || amount <= 0 {
      sendCmd.Usage()
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
      // todo
      // sendTx(knownNodes[0], tx)
    }
  }

  // if startNodeCmd.Parsed() {
  //   nodeID := os.Getenv("NODE_ID")
  //   if nodeID == "" {
  //     startNodeCmd.Usage()
  //     os.Exit(1)
  //   }
  //   fmt.Printf("Starting node %s\n", nodeID)
  //   if len(*minerAddress) > 0 {
  //     if IsAddressValid(*minerAddress) {
  //       fmt.Println("Mining is on. Address to receive rewards: ", minerAddress)
  //     } else {
  //       log.Panic("Wrong miner address!")
  //     }
  //   }
  //   StartServer(nodeID, minerAddress)
  // }
}

package main

import (
  "os"
  "fmt"
  "flag"
  "log"
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
  // cmd := make([]*flag.FlagSet, len(cmdOption))
  for _, c := range cmdOption {
    cmd = append(cmd, flag.NewFlagSet(c, flag.ExitOnError))
  }
  createBlockchainCmd := cmd[0]
  // createWalletCmd :=  cmd[1]
  // getBalanceCmd := cmd[2]
  // listAddressesCmd := cmd[3]
  // printChainCmd := cmd[4]
  // reindexUTXOCmd := cmd[5]
  // sendCmd := cmd[6]
  // startNodeCmd := cmd[7]

  // getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
  createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
  // sendFrom := sendCmd.String("from", "", "Source wallet address")
  // sendTo := sendCmd.String("to", "", "Destination wallet address")
  // sendAmount := sendCmd.Int("amount", 0, "Amount to send")
  // sendMine := sendCmd.Bool("mine", false, "Mine immediately on the same node")
  // startNodeMiner := startNodeCmd.String("miner", "", "Enable mining mode and send reward to ADDRESS")
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
    if *createBlockchainAddress == "" {
      createBlockchainCmd.Usage()
      os.Exit(1)
    }
    // if !ValidateAddress(address) {
    //   log.Panic("ERROR: Address is not valid")
    // }
    bc := createBlockchainDB(*createBlockchainAddress, nodeID)
    defer bc.db.Close()
    // UTXOSet := UTXOSet{bc}
    // UTXOSet.Reindex()

    fmt.Println("Done!")
  }

  // getBalance
  // if getBalanceCmd.Parsed() {
  //   if *getBalanceAddress == "" {
  //     getBalanceCmd.Usage()
  //     os.Exit(1)
  //   }
  //   cli.getBalance(*getBalanceAddress, nodeID)
  // }
  //
  //
  // if createWalletCmd.Parsed() {
  //   cli.createWallet(nodeID)
  // }
  //
  // if listAddressesCmd.Parsed() {
  //   cli.listAddresses(nodeID)
  // }
  //
  // if printChainCmd.Parsed() {
  //   cli.printChain(nodeID)
  // }
  //
  // if reindexUTXOCmd.Parsed() {
  //   cli.reindexUTXO(nodeID)
  // }
  //
  // if sendCmd.Parsed() {
  //   if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
  //     sendCmd.Usage()
  //     os.Exit(1)
  //   }
  //   cli.send(*sendFrom, *sendTo, *sendAmount, nodeID, *sendMine)
  // }
  //
  // if startNodeCmd.Parsed() {
  //   nodeID := os.Getenv("NODE_ID")
  //   if nodeID == "" {
  //     startNodeCmd.Usage()
  //     os.Exit(1)
  //   }
  //   cli.startNode(nodeID, *startNodeMiner)
  // }
}
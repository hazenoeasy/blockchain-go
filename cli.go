package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

type CLI struct {
	bc *BlockChain
}

func (cli *CLI) printUsage() {
	// fmt.Printf(usage)
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) createblockchain(address string) {

	if !ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := CreateBlockChain(address)
	defer bc.db.Close()
	fmt.Println("Done!")
}

func (cli *CLI) printChain() {
	bc := GetBlockchain()
	defer bc.db.Close()
	bci := bc.Iterator()
	for {
		block := bci.Iter()

		fmt.Printf("Prev hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		// fmt.Printf("Transactions: %s", block.Transactions[0])
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevBlockHash) == 0 { // get the Genesis Block
			break
		}
	}
}

func (cli *CLI) getBalance(address string) {
	bc := GetBlockchain()
	defer bc.db.Close()

	if !ValidateAddress(address) {
		fmt.Println("not a valid address")
		os.Exit(1)
	}
	balance := 0
	pubKeyHash := GetPubKeyHashfromAddress(address) // get public key hash
	UTXOs := bc.FindUTXO(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

func (cli *CLI) send(from, to string, amount int) {
	cli.bc = GetBlockchain()
	defer cli.bc.db.Close()
	var txs []*Transaction
	tx := NewUTXOTransaction(from, to, amount, cli.bc)
	txs = append(txs, tx)
	cli.bc.MineBlock(txs) // transactions 都在 block里
	fmt.Println("have send the money!")
}

func (cli *CLI) createwallet() {
	wallets := NewWallets()
	s := wallets.CreateWallet()
	wallets.SaveToFile()
	fmt.Printf("Done! your wallet address is %s\n", s)
}

func (cli *CLI) listAddresses() {
	wallets := NewWallets()
	addresses := wallets.GetAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}

func (cli *CLI) Run() {
	cli.validateArgs()

	createblockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	getbalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	createwalletCmd := flag.NewFlagSet("createwallet", flag.ContinueOnError)
	createblockchainData := createblockchainCmd.String("address", "", "Address")
	getbalanceData := getbalanceCmd.String("address", "", "Address")
	sendData1 := sendCmd.String("from", "", "From")
	sendData2 := sendCmd.String("to", "", "To")
	sendData3 := sendCmd.Int("amount", 0, "Amount")
	listAddressesCmd := flag.NewFlagSet("listAddresses", flag.ExitOnError)
	var err error
	switch os.Args[1] {

	case "printchain":
		err = printChainCmd.Parse(os.Args[2:])
	case "getbalance":
		err = getbalanceCmd.Parse(os.Args[2:])
	case "createblockchain":
		err = createblockchainCmd.Parse(os.Args[2:])
	case "send":
		err = sendCmd.Parse(os.Args[2:])
	case "createwallet":
		err = createwalletCmd.Parse(os.Args[2:])
	case "listaddress":
		err = listAddressesCmd.Parse(os.Args[2:])
	default:
		cli.printUsage()
		os.Exit(1)
	}
	if err != nil {
		log.Panic(err)
	}

	if createblockchainCmd.Parsed() {
		if *createblockchainData == "" {
			createblockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createblockchain(*createblockchainData)
	}

	if getbalanceCmd.Parsed() {
		if *getbalanceData == "" {
			getbalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getbalanceData)
	}
	if sendCmd.Parsed() {
		if *sendData1 == "" {
			sendCmd.Usage()
			os.Exit(1)
		}
		if *sendData2 == "" {
			sendCmd.Usage()
			os.Exit(1)
		}
		if *sendData3 == 0 {
			sendCmd.Usage()
			os.Exit(1)
		}
		cli.send(*sendData1, *sendData2, *sendData3)
	}
	if printChainCmd.Parsed() {
		cli.printChain()
	}
	if createwalletCmd.Parsed() {
		cli.createwallet()
	}
	if listAddressesCmd.Parsed() {
		cli.listAddresses()
	}
}

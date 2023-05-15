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

const usage = `
Usage:
	createblockchain	-address 	Address		create a blockchain
	getbalance 			-address 	Address		get balance
	printchain                   				print all the blocks of the blockchain
	send				-from 		Address		-to 		Address 		-amount 	value	send the money from A to B
`

func (cli *CLI) printUsage() {
	fmt.Println(usage)
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) createblockchain(address string) {
	bc := NewBlockChain(address)
	bc.db.Close()
	fmt.Println("Done!")
}

func (cli *CLI) printChain() {
	cli.bc = NewBlockChain("") // if there was a blockchain in the db, it will read the l block
	bci := cli.bc.Iterator()
	for {
		block := bci.Iter()

		fmt.Printf("Prev hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevBlockHash) == 0 { // get the Genesis Block
			break
		}
	}
}

func (cli *CLI) getBalance(address string) {
	bc := NewBlockChain(address)
	defer bc.db.Close()

	balance := 0
	UTXOs := bc.FindUTXO(address)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

func (cli *CLI) send(from, to string, amount int) {
	cli.bc = NewBlockChain(from)
	defer cli.bc.db.Close()
	var txs []*Transaction
	tx := NewUTXOTransaction(from, to, amount, cli.bc)
	txs = append(txs, tx)
	cli.bc.MineBlock(txs)
	fmt.Println("have send the money!")

}
func (cli *CLI) Run() {
	cli.validateArgs()

	createblockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	getbalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)

	createblockchainData := createblockchainCmd.String("address", "", "Address")
	getbalanceData := getbalanceCmd.String("address", "", "Address")
	sendData1 := sendCmd.String("from", "", "From")
	sendData2 := sendCmd.String("to", "", "To")
	sendData3 := sendCmd.Int("amount", 0, "Amount")

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
}

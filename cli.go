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
  addblock -data BLOCK_DATA    add a block to the blockchain
  printchain                   print all the blocks of the blockchain
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
func (cli *CLI) Run() {
	cli.validateArgs()

	createblockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	createblockchainData := createblockchainCmd.String("address", "", "Address")
	var err error
	switch os.Args[1] {

	case "printchain":
		err = printChainCmd.Parse(os.Args[2:])
	case "createblockchain":
		err = createblockchainCmd.Parse(os.Args[2:])
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

	if printChainCmd.Parsed() {
		cli.printChain()
	}
}

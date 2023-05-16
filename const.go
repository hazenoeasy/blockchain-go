package main

import "math"

const dbFile = "blockchain.db"
const utxoBucket = "utxo"
const blocksBucket = "blocks"
const usage = `Usage:
	createblockchain	-address 	Address		create a blockchain
	getbalance 			-address 	Address		get balance
	printchain                   				print all the blocks of the blockchain
	send				-from 		Address		-to 		Address 		-amount 	value	send the money from A to B
`

// difficulty
const targetBits = 1
const maxNonce = math.MaxInt32
const subsidy = 10

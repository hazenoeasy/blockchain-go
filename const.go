package main

import "math"

const dbFile = "blockchain.db"
const blocksBucket = "blocks"
const usage = `
Usage:
	createblockchain	-address 	Address		create a blockchain
	getbalance 			-address 	Address		get balance
	printchain                   				print all the blocks of the blockchain
	send				-from 		Address		-to 		Address 		-amount 	value	send the money from A to B
`

// difficulty
const targetBits = 24
const maxNonce = math.MaxInt32
const subsidy = 10

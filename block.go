package main

import (
	"bytes"
	"crypto/sha256"
	"strconv"
	"time"
)

// define the Block structure
type Block struct {
	Timestamp     int64  // timestamp of the create time
	Data          []byte // stored Data
	PrevBlockHash []byte // previous block's hash
	Hash          []byte // hash of the above information
}

// calculate the hash value of the current block
func (b *Block) SetHash() { // instance method
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))                       // transfer int64 into slice
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{}) //  []byte{} is an empty slice, it will be the separator for the concat header
	hash := sha256.Sum256(headers)                                                // fixed-length array of 32 bytes.

	b.Hash = hash[:] // create a slice from an array
}

// create a block
func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}}
	block.SetHash()
	return block
}

func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}

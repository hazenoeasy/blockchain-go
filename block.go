package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"strconv"
	"time"
)

// define the Block structure
type Block struct {
	Timestamp     int64  // timestamp of the create time
	Data          []byte // stored Data
	PrevBlockHash []byte // previous block's hash
	Hash          []byte // hash of the above information
	Nonce         int
}

// calculate the hash value of the current block
func (b *Block) SetHash() { // instance method
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))                       // transfer int64 into slice
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{}) //  []byte{} is an empty slice, it will be the separator for the concat header
	hash := sha256.Sum256(headers)                                                // fixed-length array of 32 bytes.

	b.Hash = hash[:] // create a slice from an array
}
func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	encoder.Encode(b) // need to handle the error

	return result.Bytes()
}
func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	decoder.Decode(&block) // need to handle the error

	return &block
}

// create a block
func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}, 0}
	pof := NewProofOfWork(block)
	nonce, hash := pof.Run()
	block.Hash = hash
	block.Nonce = nonce
	return block
}

func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}

package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

// define the Block structure
type Block struct {
	Timestamp     int64 // timestamp of the create time
	Transactions  []*Transaction
	PrevBlockHash []byte // previous block's hash
	Hash          []byte // hash of the above information
	Nonce         int
}

func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b) // need to handle the error
	if err != nil {
		log.Panic("encode issue!")
	}

	return result.Bytes()
}
func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block) // need to handle the error
	if err != nil {
		log.Panic("decode issue!")
	}

	return &block
}

// create a block
func NewBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), transactions, prevBlockHash, []byte{}, 0}
	pof := NewProofOfWork(block)
	nonce, hash := pof.Run()
	block.Hash = hash
	block.Nonce = nonce
	return block
}

func (b *Block) HashTransactions() []byte { // get the  value of the rootNode
	var transactions [][]byte

	for _, tx := range b.Transactions {
		transactions = append(transactions, tx.Serialize())
	}

	mTree := NewMerkleTree(transactions)

	return mTree.RootNode.Data
}

func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

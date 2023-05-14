package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
)

const subsidy = 10

// 一笔交易会有多个input和多个output
type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

// 一个input 由Txid表明是哪个transaction 从而找到是哪个output
type TXInput struct {
	Txid      []byte
	Vout      int    //  an index of an output in the transaction.
	ScriptSig string // if the data is correct, the output can be unlocked.
}

// output 记录了scriptPubkey 还有有的value
type TXOutput struct {
	Value        int
	ScriptPubKey string
}

// SetID sets ID of a transaction
func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx) // encode tx into bytes
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

// func NewUTXOTransaction(from string, to string, amount int, bc *Block) TXOutput {

// }
func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txin := TXInput{[]byte{}, -1, data}                        // Transaction Id is null, index should be -1
	txout := TXOutput{subsidy, to}                             // value is the reward, scriptPubKey is the target address
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}} // no transaction ID
	tx.SetID()

	return &tx
}

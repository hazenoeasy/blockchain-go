package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"log"
)

// 一笔交易会有多个input和多个output
type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
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

func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

func NewUTXOTransaction(from, to string, amount int, bc *BlockChain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	// check if from and to is legal
	if !ValidateAddress(from) {
		log.Panic("illegal address!")
	}
	if !ValidateAddress(to) {
		log.Panic("illegal address!")
	}
	wallets := NewWallets()
	wallet := wallets.GetWallet(from)
	//get pubKeyHash
	pubKeyHash := GetPubKeyHashfromAddress(from)
	// get acc and validOutputs
	acc, validOutputs := bc.FindSpendableOutputs(pubKeyHash, amount)

	if acc < amount {
		log.Panic("ERROR: Not enough funds")
	}

	// Build a list of inputs
	for txid, outs := range validOutputs { // 可以使用的output
		txID, _ := hex.DecodeString(txid) // string to bytes

		for _, out := range outs {
			input := TXInput{txID, out, []byte{}, wallet.PublicKey} // 构建为新的input
			inputs = append(inputs, input)
		}
	}

	// Build a list of outputs // output 改成就两个 一个是目标的 一个是剩余的
	outputs = append(outputs, *NewTXOutput(amount, to))
	if acc > amount {
		outputs = append(outputs, *NewTXOutput(acc-amount, from)) // a change
	}

	tx := Transaction{nil, inputs, outputs}
	tx.SetID()

	return &tx
}

func NewCoinbaseTX(to string) *Transaction {
	txin := TXInput{[]byte{}, -1, nil, []byte("New Coin Base ")} // Transaction Id is null, index should be -1
	txout := NewTXOutput(subsidy, to)                            // value is the reward, scriptPubKey is the target address
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}}  // no transaction ID
	tx.SetID()

	return &tx
}
func GetPubKeyHashfromAddress(address string) []byte {
	pubKeyHash := Base58Decode([]byte(address))    //将address进行解码
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4] // 去掉version 还有checksum
	return pubKeyHash
}

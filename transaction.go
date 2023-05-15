package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
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

func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

func (in *TXInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}

func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}

func (bc *BlockChain) FindUnspentTransactions(address string) []Transaction {
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int) //map  string : int[]切片 一个transaction中 已经花掉的TXO
	bci := bc.Iterator()

	for {
		block := bci.Iter() // the last block

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID) // transfer id to string

		Outputs:
			for outIdx, out := range tx.Vout { // iter out put  outIdx is the index , out is the TxOutput
				// Was the output spent?
				if spentTXOs[txID] != nil { // 如果关于这个transaction 有记录
					for _, spentOut := range spentTXOs[txID] { // 遍历 transaction 的切片
						if spentOut == outIdx { // 该TXO已经被花了 所以直接返回
							continue Outputs
						}
					}
				}
				// 该TXO还没花
				if out.CanBeUnlockedWith(address) { // 证明这个out 的value 属于这个address， out是从这里出去的
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			if !tx.IsCoinbase() { // 不是初始块
				for _, in := range tx.Vin { // 查看 Vin
					if in.CanUnlockOutputWith(address) { // input 可以被解锁， input与这个address相关 input是指向这个address的
						inTxID := hex.EncodeToString(in.Txid)                  //  transaction的 ID
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout) // 指向input的output是已经使用的value 所以应该加入spendTXO中
					}
				}
			}
		}

		if len(block.PrevBlockHash) == 0 { // 如果到头了 就停下来
			break
		}
	}

	return unspentTXs
}

func (bc *BlockChain) FindUTXO(address string) []TXOutput { // 可以合并
	var UTXOs []TXOutput
	unspentTransactions := bc.FindUnspentTransactions(address)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

func NewUTXOTransaction(from, to string, amount int, bc *BlockChain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	acc, validOutputs := bc.FindSpendableOutputs(from, amount)

	if acc < amount {
		log.Panic("ERROR: Not enough funds")
	}

	// Build a list of inputs
	for txid, outs := range validOutputs { // 可以使用的output
		txID, _ := hex.DecodeString(txid) // string to bytes

		for _, out := range outs {
			input := TXInput{txID, out, from} // 构建为新的input
			inputs = append(inputs, input)
		}
	}

	// Build a list of outputs // output 改成就两个 一个是目标的 一个是剩余的
	outputs = append(outputs, TXOutput{amount, to})
	if acc > amount {
		outputs = append(outputs, TXOutput{acc - amount, from}) // a change
	}

	tx := Transaction{nil, inputs, outputs}
	tx.SetID()

	return &tx
}

func (bc *BlockChain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(address) //找到还没花完的
	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout { // 遍历 out
			if out.CanBeUnlockedWith(address) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOutputs
}

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

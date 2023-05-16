package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
)

type BlockChain struct {
	tip []byte
	db  *bolt.DB
}

func (bc *BlockChain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db}

	return bci
}

func (bc *BlockChain) MineBlock(transaction []*Transaction) *Block {
	var lastHash []byte
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	for _, tx := range transaction {
		if !bc.VerifyTransaction(tx) {
			fmt.Printf("%x transaction didn't pass the verification\n", tx.ID)
			log.Panic("ERROR: Invalid transaction")
		}
	}
	newBlock := NewBlock(transaction, lastHash) // create new block
	// add the new block to the database
	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		b.Put(newBlock.Hash, newBlock.Serialize())
		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}
		bc.tip = newBlock.Hash
		return nil
	})
	fmt.Printf("Mine the block  %x\n", newBlock.Hash)
	return newBlock
}

func DeleteDB() {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		fmt.Printf("dbFile doesn't exist\n")
		return
	}
	err := os.Remove(dbFile)
	if err != nil {
		fmt.Printf("Failed to delete dbFile: %v\n", err)
		return
	}

	fmt.Println("File deleted successfully.")
}
func CreateBlockChain(address string) *BlockChain {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}
	var tip []byte                          // last one
	db, err := bolt.Open(dbFile, 0600, nil) // need to handle the error
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		cbtx := NewCoinbaseTX(address)
		genesis := NewGenesisBlock(cbtx)
		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			return err
		}
		err = b.Put(genesis.Hash, genesis.Serialize())
		if err != nil {
			return err
		}
		err = b.Put([]byte("l"), genesis.Hash) // l is pointed to the last block
		if err != nil {
			return err
		}
		tip = genesis.Hash

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return &BlockChain{tip, db}
}

// Get Blockchain
func GetBlockchain() *BlockChain {
	if dbExists() == false {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	bc := BlockChain{tip, db}

	return &bc
}

// 效率太低 每次都需要遍历一遍
func (bc *BlockChain) FindUnspentTransactions_Old(pubKeyHash []byte) []Transaction {
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int) //map  string : int[]切片 一个transaction中 已经花掉的TXO
	bci := bc.Iterator()

	for {
		block := bci.Iter() // the last block

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID) // transfer id to string
			// 要是自己转自己转帐呢？
			// func (tx Transaction) IsCoinbase() bool {
			// 	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
			// }
			if !tx.IsCoinbase() { // 不是初始块 那么就会有input
				for _, in := range tx.Vin { // 查看 Vin
					if in.UsesKey(pubKeyHash) { //  判断这个input 是 该public key发出的, 说明这个input 对应的output已经被用了
						inTxID := hex.EncodeToString(in.Txid)                  //  transaction的 ID
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout) // 指向input的output是已经使用的value 所以应该加入spendTXO中
					}
				}
			}
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
				if out.IsLockedWithKey(pubKeyHash) { // 证明这个out 的value 属于这个address
					unspentTXs = append(unspentTXs, *tx)
				}
			}

		}

		if len(block.PrevBlockHash) == 0 { // 如果到头了 就停下来
			break
		}
	}

	return unspentTXs
}

// 怎么提高效率呢？
// 有一个index保存unspent output  -> UTXO set
// UTXO is a cache that is built from all blockchain transactions, and it later used to calculate balance and validate new transactions.
// 有了FindUTXO后， 就不需要这个啦
// func (bc *BlockChain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
// 	var unspentTXs []Transaction
// 	spentTXOs := make(map[string][]int) //map  string : int[]切片 一个transaction中 已经花掉的TXO
// 	bci := bc.Iterator()

// 	for {
// 		block := bci.Iter() // the last block

// 		for _, tx := range block.Transactions {
// 			txID := hex.EncodeToString(tx.ID) // transfer id to string
// 			// 要是自己转自己转帐呢？
// 			// func (tx Transaction) IsCoinbase() bool {
// 			// 	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
// 			// }
// 			if !tx.IsCoinbase() { // 不是初始块 那么就会有input
// 				for _, in := range tx.Vin { // 查看 Vin
// 					if in.UsesKey(pubKeyHash) { //  判断这个input 是 该public key发出的, 说明这个input 对应的output已经被用了
// 						inTxID := hex.EncodeToString(in.Txid)                  //  transaction的 ID
// 						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout) // 指向input的output是已经使用的value 所以应该加入spendTXO中
// 					}
// 				}
// 			}
// 		Outputs:
// 			for outIdx, out := range tx.Vout { // iter out put  outIdx is the index , out is the TxOutput
// 				// Was the output spent?
// 				if spentTXOs[txID] != nil { // 如果关于这个transaction 有记录
// 					for _, spentOut := range spentTXOs[txID] { // 遍历 transaction 的切片
// 						if spentOut == outIdx { // 该TXO已经被花了 所以直接返回
// 							continue Outputs
// 						}
// 					}
// 				}
// 				// 该TXO还没花
// 				if out.IsLockedWithKey(pubKeyHash) { // 证明这个out 的value 属于这个address
// 					unspentTXs = append(unspentTXs, *tx)
// 				}
// 			}

// 		}

// 		if len(block.PrevBlockHash) == 0 { // 如果到头了 就停下来
// 			break
// 		}
// 	}

// 	return unspentTXs
// }

// – finds all unspent outputs by iterating over blocks.
func (bc *BlockChain) FindUTXO() map[string]TXOutputs { // 可以合并
	UTXO := make(map[string]TXOutputs)
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()
	for {
		block := bci.Iter()
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)
		Outputs:
			for outIdx, out := range tx.Vout {
				if spentTXOs[txID] != nil {
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx { // 这些index的 是已经用过的 这种实现方法也不太好
							continue Outputs
						}
					}
				}
				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				outs.OutIdxs = append(outs.OutIdxs, outIdx)
				UTXO[txID] = outs

			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.Txid)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
				}
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}

	}
	return UTXO
}

// 根据transaction ID 搜索得到Transaction
func (bc *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()
	for {
		block := bci.Iter()

		for _, tx := range block.Transactions {
			if bytes.Equal(tx.ID, ID) {
				return *tx, nil
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return Transaction{}, errors.New("Transaction is not found")
}

// 对Transaction签名
func (bc *BlockChain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)     // 根据 vin的 transaction 找到来源的transaction
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX // id和transaction 对应
		if err != nil {
			log.Panic(err)
		}
	}

	tx.Sign(privKey, prevTXs)
}

func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.Txid)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
		if err != nil {
			log.Panic(err)
		}
	}
	return tx.Verify(prevTXs)
}

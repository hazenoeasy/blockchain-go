package main

import (
	"encoding/hex"
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

func (bc *BlockChain) MineBlock(transaction []*Transaction) {
	var lastHash []byte
	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panic(err)
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
		fmt.Printf("%s", genesis.Serialize())
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

func (bc *BlockChain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(pubKeyHash) //找到还没花完的
	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout { // 遍历 out
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
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

func (bc *BlockChain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
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

func (bc *BlockChain) FindUTXO(pubKeyHash []byte) []TXOutput { // 可以合并
	var UTXOs []TXOutput
	unspentTransactions := bc.FindUnspentTransactions(pubKeyHash)

	for _, tx := range unspentTransactions { // 都是可用的transaction的out, 必须这么记载么？
		for _, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

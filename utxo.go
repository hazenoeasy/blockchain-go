package main

import (
	"encoding/hex"
	"log"

	"github.com/boltdb/bolt"
)

// store the UTXO set in a different UTXO set
type UTXOSet struct {
	BlockChain *BlockChain
}

// uses FindUTXO to find unspent outputs, and stores them in a database. This is where caching happens.
func (u UTXOSet) Reindex() {
	db := u.BlockChain.db
	bucketName := []byte(utxoBucket)
	err := db.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket(bucketName)
		_, err := tx.CreateBucket(bucketName)
		return err
	})
	if err != nil {
		log.Panic(err)
	}
	// find all unspent UTXO  找到每个transaction中对应的未使用的out 数组
	UTXO := u.BlockChain.FindUTXO()
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		for txID, outs := range UTXO {
			key, err := hex.DecodeString(txID)
			if err != nil {
				return err
			}
			err = b.Put(key, outs.Serialize())
		}
		return err
	})
}

func (uxto *UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	accumulated := 0
	db := uxto.BlockChain.db
	bucketName := []byte(utxoBucket)
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			txID := hex.EncodeToString(k)
			outs := DeserializeTXOutputs(v)

			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOutputs[txID] = append(unspentOutputs[txID], outs.OutIdxs[outIdx])
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return accumulated, unspentOutputs
}

// 从数据库中读出pubKeyHash 的TXOutputs
func (uxto *UTXOSet) FindUTXO(pubKeyHash []byte) []TXOutput {
	var UTXOs []TXOutput
	db := uxto.BlockChain.db
	bucketName := []byte(utxoBucket)
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs := DeserializeTXOutputs(v)
			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	return UTXOs
}

// 更新 UTXOSet数据库
// 1 遍历block中的transactions
// 2 如果是查询in 对应的 transaction， 并在UTXOSet中删除
// 3 如果是out 则加入UTXOSet
func (u UTXOSet) Update(block *Block) {
	db := u.BlockChain.db
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		for _, tx := range block.Transactions { // 遍历 transactions
			// 处理in
			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					updateOuts := TXOutputs{}
					outsBytes := b.Get(in.Txid) // 得到 transaction 对应的没用过的outputs
					outs := DeserializeTXOutputs(outsBytes)
					for idx, outIdx := range outs.OutIdxs {
						if outIdx != in.Vout { // 证明这个还没用过
							updateOuts.Outputs = append(updateOuts.Outputs, outs.Outputs[idx]) // 把数据加进去
							updateOuts.OutIdxs = append(updateOuts.OutIdxs, outIdx)            // 把index加进去
						}
					}
					if len(updateOuts.Outputs) == 0 { // 一个合格的都没有
						err := b.Delete(in.Txid)
						if err != nil {
							return err
						}
					} else { // 更新
						err := b.Put(in.Txid, updateOuts.Serialize())
						if err != nil {
							return err
						}
					}

				}
			}
			// 处理 out
			newOutputs := TXOutputs{}
			for idx, out := range tx.Vout {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
				newOutputs.OutIdxs = append(newOutputs.OutIdxs, idx)
			}

			err := b.Put(tx.ID, newOutputs.Serialize())
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Panic()
	}
}

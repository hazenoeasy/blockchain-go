package main

import "bytes"

// 一个input 由Txid表明是哪个transaction 从而找到是哪个output
type TXInput struct { // input 说明了 钱是哪里来的  那这个也可以随便伪造嘛？ 肯定不行，只有数据的来源，有私钥，可以对数据进行签名， 接收方通过公钥来解密
	Txid []byte // transaction ID
	Vout int    //  an index of an output in the transaction.
	// ScriptSig string // if the data is correct, the output can be unlocked.
	Signature []byte // the signature of the content
	PubKey    []byte // the public key who send this money
}

// UsesKey checks whether the address initiated the transaction 检查 是否是这个地址的人 发出了input
func (in *TXInput) UsesKey(pubKeyHash []byte) bool { // 对input的public key 进行两次hash， 得到的值为加锁值，跟pubKeyHash进行对比
	lockingHash := HashPubKey(in.PubKey)
	return bytes.Compare(lockingHash, pubKeyHash) == 0
}

package main

import "bytes"

// 一个input 由Txid表明是哪个transaction 从而找到是哪个output
type TXInput struct { // input 说明了 钱是哪里来的  那这个也可以随便伪造嘛？ 肯定不行，只有数据的来源，有私钥，可以对数据进行签名， 接收方通过公钥来解密
	Txid []byte // 先前的交易的哈希 transaction ID  来源的transaction的id
	Vout int    //  先前交易的输出索引 an index of an output in the transaction. 这个input 是由来源的transaction中哪个output构成的
	// ScriptSig string // if the data is correct, the output can be unlocked.
	Signature []byte // the signature of the content
	PubKey    []byte // the public key who send this money, public key 经过HashPubKey 得到pubKeyHash, 经过封装可以得到address 所以address 也可以逆向得到pubKeyHash
	// 说白了就是来源output的pub key 来表明身份
}

// UsesKey checks whether the address initiated the transaction 检查 是否是这个地址的人 发出了input
func (in *TXInput) UsesKey(pubKeyHash []byte) bool { // 对input的public key 进行两次hash， 得到的值为加锁值，跟pubKeyHash进行对比
	lockingHash := HashPubKey(in.PubKey)
	return bytes.Compare(lockingHash, pubKeyHash) == 0
}

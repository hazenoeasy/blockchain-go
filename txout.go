package main

import "bytes"

type TXOutput struct { // 遍历output 才能知道balance  pubKeyHash表明了钱流向了谁，这个也可以伪造吗， 那肯定不行，要不然我都写我自己了。
	Value int
	// ScriptPubKey string
	PubKeyHash []byte // claim that this output belongs to someone
}

// Lock signs the output   // 通过目标的Address来获得pubKeyHash 并加到output上面
func (out *TXOutput) Lock(address []byte) {
	pubKeyHash := Base58Decode(address)            //将address进行解码
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4] // 去掉version 还有checksum
	out.PubKeyHash = pubKeyHash                    // 将adderss中的公共密钥hash值提取出来，并计入out中
}

// IsLockedWithKey checks if the output can be used by the owner of the pubkey
func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool { // 判断out是不是由该public key lock
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

func NewTXOutput(value int, address string) *TXOutput {
	tx := &TXOutput{value, nil}
	tx.Lock([]byte(address))

	return tx
}

package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"log"
	"math/big"
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
	bc.SignTransaction(&tx, wallet.PrivateKey)
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

// serialize the transaction
func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

// Hash returns the hash of the Transaction
func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{} // set ID is empty 不能暴露这个ID ？

	hash = sha256.Sum256(txCopy.Serialize()) // 256 来 hash

	return hash[:]
}

// Trim signature and pub key
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput
	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.Txid, vin.Vout, nil, nil})
	}
	copy(tx.Vout, outputs)
	txCopy := Transaction{tx.ID, inputs, outputs}
	return txCopy

}

// 签名输入：发送方使用其私钥对交易输入进行签名。
// 这个签名的目的是证明发送方拥有与之关联的公钥的所有权。签名过程通常涉及使用椭圆曲线数字签名算法（ECDSA）。
// sign the transaction with previous transaction
// 接收private key 还有previous map，来给transaction签字
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {

	if tx.IsCoinbase() {
		return // 初始coinbase 不需要签名，因为没有发送方
	}
	txCopy := tx.TrimmedCopy()
	// 提前将prevTXs计算好 并存入map中，这样 就方便查找了
	for inID, vin := range txCopy.Vin {
		prevTXs := prevTXs[hex.EncodeToString(vin.Txid)]            // 通过transaction ID 找到 对应的transaction
		txCopy.Vin[inID].Signature = nil                            // 多重保险
		txCopy.Vin[inID].PubKey = prevTXs.Vout[vin.Vout].PubKeyHash // 找到input 对应的来源的output 将output 的公钥给到这个input
		txCopy.ID = txCopy.Hash()                                   // 将 这个ID用来短暂的存储hash值 这个hash值对这个input是独特的
		txCopy.Vin[inID].PubKey = nil                               // hash完了 重新吧这个referenced output 的公钥消除

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID) //使用privatekey 对hash进行sign
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Vin[inID].Signature = signature // 将signature写到 transaction 中
	}
}

// 验证签名：当其他节点收到交易后，它们会验证发送方的签名。
// 验证的过程包括使用发送方的公钥和签名来验证签名的有效性，以及检查交易引用的输出是否确实属于发送方。
// 使用发送方的公钥来解锁签名， 然后跟transaction的数据做对比, 同时检查引用的输出是否属于发送方
// 问题是 为什么需要prevTXs
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	// 遍历Vin
	for inID, vin := range tx.Vin {
		prevTX := prevTXs[hex.EncodeToString(vin.Txid)] // 找到前面的transaction
		// during verification we need the same data what was signed.
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTX.Vout[vin.Vout].PubKeyHash
		txCopy.ID = txCopy.Hash() //计算一下hash
		txCopy.Vin[inID].PubKey = nil

		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PubKey)
		x.SetBytes(vin.PubKey[:(keyLen / 2)])
		y.SetBytes(vin.PubKey[(keyLen / 2):])
		rawPubKey := ecdsa.PublicKey{curve, &x, &y}               // 将pubKey的value与curve合并， 然后得到rawPubKey
		if ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) == false { // 验证这个public key 能不能解锁txCopy.ID
			return false
		}
	}

	return true
}

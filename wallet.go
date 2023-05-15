package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"log"

	"golang.org/x/crypto/ripemd160"
)

const version = byte(0x00)
const walletFile = "wallet.dat"
const addressChecksumLen = 4

// struct
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte // 在实际使用中，还是使用了pubKeyHash, 因为这样可以隐藏公钥本身
}

// methods
func (w Wallet) GetAddress() []byte { // transfer from public key to bitcoin Address address 可以理解为pubKeyHash经过Base58转化为人类可读的地址
	pubKeyHash := HashPubKey(w.PublicKey)                      // RIPEMD160(SHA256(PubKey)) double hash
	versionedPayload := append([]byte{version}, pubKeyHash...) // 将verison添加到开头
	checksum := checksum(versionedPayload)                     // 生成checksum 用于校验
	fullPayload := append(versionedPayload, checksum...)       // 将校验和追加到末尾
	address := Base58Encode(fullPayload)
	return address
}

// class
func NewWallet() *Wallet {
	private, public := newKeyPair()
	return &Wallet{private, public}
}

// functions
// 将public key 转化为 pubKeyHash 来隐藏public key 本身
func HashPubKey(pubKey []byte) []byte {
	publicSHA256 := sha256.Sum256(pubKey) // 返回一个长度为32的字节数组 代表哈希结果

	RIPEMD160HASHER := ripemd160.New()               // 创建了RIPEMD-160 的实例
	_, err := RIPEMD160HASHER.Write(publicSHA256[:]) // 将hash结果写入实例
	if err != nil {
		log.Panic(err)
	}
	publicRIPEMD160 := RIPEMD160HASHER.Sum(nil) //Sum 方法接收一个字节切片作为输入，这里传递nil表示返回完整的哈希结果。
	return publicRIPEMD160
}

func checksum(payload []byte) []byte { // 生成一个校验和
	firstSHA := sha256.Sum256(payload)      // 使用SHA256对payload进行hash计算 返回长度为32的字节数组
	secondSHA := sha256.Sum256(firstSHA[:]) //二次hash

	return secondSHA[:addressChecksumLen] //返回前几位作为校验和
}

func ValidateAddress(address string) bool { // 再算一遍， 要是check sum符合 就是正确的
	pubKeyHash := Base58Decode([]byte(address))
	actualChecksum := pubKeyHash[len(pubKeyHash)-addressChecksumLen:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-addressChecksumLen]
	targetChecksum := checksum(append([]byte{version}, pubKeyHash...))

	return bytes.Compare(actualChecksum, targetChecksum) == 0
}

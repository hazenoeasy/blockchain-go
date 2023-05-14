package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

// difficulty
const targetBits = 24
const maxNonce = math.MaxInt32

type ProofOfWork struct {
	block  *Block
	target *big.Int
}

func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits)) // left shit, so the left targetBits will be zero
	pow := &ProofOfWork{b, target}
	return pow
}

func (pow *ProofOfWork) generateBytes(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.HashTransactions(),
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)

	return data
}

func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte // 32 bytes hash value
	nonce := 0

	fmt.Printf("Mining the block containing \"%s\"\n", pow.block.Hash)
	for nonce < maxNonce {
		data := pow.generateBytes(nonce)
		hash = sha256.Sum256(data)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			fmt.Printf("\r%x", hash)
			break
		} else {
			nonce++
		}
	}
	fmt.Print("\n\n")

	return nonce, hash[:] // change the type
}

func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.generateBytes(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}

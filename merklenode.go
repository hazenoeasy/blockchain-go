package main

import "crypto/sha256"

type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data  []byte
}

func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	mNode := MerkleNode{}

	if left == nil && right == nil { // 如果左右都为空， 就是叶子结点
		hash := sha256.Sum256(data)
		mNode.Data = hash[:]
	} else { //
		prevHashes := append(left.Data, right.Data...) // 将左右的hash值拼起来，然后再hash一遍
		hash := sha256.Sum256(prevHashes)
		mNode.Data = hash[:]
	}

	mNode.Left = left
	mNode.Right = right

	return &mNode
}

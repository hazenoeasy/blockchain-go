package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/binary"
	"log"
	"os"
)

// 将一个 int64 转化为一个字节数组(byte array)
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}
func newKeyPair() (ecdsa.PrivateKey, []byte) {
	// ECDSA is based on elliptic curves
	// in the elliptic curves algorithm, public keys are points on a curve, is a combination of X,Y coordinates.
	curve := elliptic.P256()
	private, _ := ecdsa.GenerateKey(curve, rand.Reader)                              // use ecdsa to generate a private key
	publicKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...) // use private key to generate public key
	return *private, publicKey
}

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

// ReverseBytes reverses a byte array
func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}

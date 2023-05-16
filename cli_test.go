package main

import (
	"fmt"
	"os"
	"testing"
)

var tom string
var jack string
var cli CLI

func setup() {
	fmt.Println("prepare works...")
	cli = CLI{}
	DeleteWallets()
	DeleteDB()
	wallets := NewWallets()
	tom = wallets.CreateWallet()
	jack = wallets.CreateWallet()
	wallets.SaveToFile()
	fmt.Println("print all wallet/address")
	cli.listAddresses()
}
func TestCreateblockchain(t *testing.T) {
	fmt.Println("create the blockchain with the start of tom")
	cli.createblockchain(tom)
}

func TestSend(t *testing.T) {
	fmt.Println("testing send 1/2/3 from tom to jack")
	cli.send(tom, jack, 1)
	cli.send(tom, jack, 2)
	// cli.send(tom, jack, 3)

}
func TestGetBalance(t *testing.T) {
	fmt.Println("print tom's balance")
	cli.getBalance(tom)
	fmt.Println("print jack's balance")
	cli.getBalance(jack)
}
func TestPrintChain(t *testing.T) {
	fmt.Println("print the who chain")
	cli.printChain()

}
func TestDebug(t *testing.T) {
	setup()
	TestCreateblockchain(t)
	TestSend(t)
	TestGetBalance(t)

}
func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

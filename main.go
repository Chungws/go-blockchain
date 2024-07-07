package main

import (
	"github.com/Chungws/go-blockchain/block"
)

func main() {
	bc := block.NewBlockChain()

	bc.AddBlock("Send 1 BTC to Ivan")
	bc.AddBlock("Send 2 more BTC to Ivan")
}

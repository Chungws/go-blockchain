package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

const utxoBucket = "chainstate"

type CLI struct {
	bc *BlockChain
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  printchain - print all the blocks of the blockchain")
	fmt.Println("  createblockchain -address ADDRESS - create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  createwallet - generates a new key-pair and saves it into the wallet file")
	fmt.Println("  listaddresses - lists all addresses from the wallet file")
	fmt.Println("  getbalance -address ADDRESS - get balance of ADDRESS")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT - send AMOUNT of coins from FROM address to TO")
}

func (cli *CLI) Run() {
	cli.validateArgs()

	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	createChainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)

	createChainAddress := createChainCmd.String("address", "", "The address to send genesis block reward to")
	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	switch os.Args[1] {
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panicln("printchain command parse failed!: ", err)
		}
	case "createblockchain":
		err := createChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic("createblockchain command parse failed!: ", err)
		}
	case "createwallet":
		err := createChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic("createwallet command parse failed!: ", err)
		}
	case "listaddresses":
		err := listAddressesCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic("listaddresses command parse failed!: ", err)
		}
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic("getbalance command parse failed!: ", err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic("send command parse failed!: ", err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if createChainCmd.Parsed() {
		if *createChainAddress == "" {
			createChainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockChain(*createChainAddress)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet()
	}

	if listAddressesCmd.Parsed() {
		cli.listAddresses()
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress)
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount == 0 {
			sendCmd.Usage()
			os.Exit(1)
		}

		cli.send(*sendFrom, *sendTo, *sendAmount)
	}
}

func (cli *CLI) printChain() {
	bci := cli.bc.Iterator()

	for {
		b := bci.Next()
		fmt.Printf("Prev. hash: %x\n", b.PrevBlockHash)
		fmt.Printf("Hash: %x\n", b.Hash)
		pow := NewProofOfWork(b)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(b.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) createBlockChain(address string) {
	bc := CreateBlockChain(address)
	defer bc.db.Close()

	UTXOSet := UTXOSet{bc}
	UTXOSet.Reindex()

	fmt.Println("Done!")
}

func (cli *CLI) getBalance(address string) {
	bc := NewBlockChain(address)
	UTXOSet := UTXOSet{bc}
	defer bc.db.Close()

	balance := 0
	pubKeyHash := Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUTXO(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s' : %d\n", address, balance)
}

func (cli *CLI) send(from, to string, amount int) {
	bc := NewBlockChain(from)
	UTXOSet := UTXOSet{bc}
	defer bc.db.Close()

	tx := NewUTXOTransaction(from, to, amount, &UTXOSet)
	cbTx := NewCoinbaseTX(from, "")
	txs := []*Transaction{cbTx, tx}

	newBlock := bc.MineBlock(txs)
	UTXOSet.Update(newBlock)
	fmt.Println("Success!")
}

func (cli *CLI) createWallet() {
	wallets, _ := NewWallets()
	address := wallets.CreateWallet()
	wallets.SaveToFile()

	fmt.Printf("Your new address: %s\n", address)
}

func (cli *CLI) listAddresses() {
	wallets, err := NewWallets()
	if err != nil {
		log.Panic(err)
	}
	addresses := wallets.GetAddresses()
	for _, address := range addresses {
		fmt.Println(address)
	}
}

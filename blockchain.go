package main

import (
	"log"

	"github.com/boltdb/bolt"
)

const (
	dbFile       = "./database.db"
	blocksBucket = "blocks"
)

type BlockChain struct {
	tip []byte
	db  *bolt.DB
}

func NewBlockChain() *BlockChain {
	var tip []byte

	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Fatalln("Failed to open blockchain db: ", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		if b == nil {
			genesis := NewGenesisBlock()
			b, err := tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				return err
			}
			err = b.Put(genesis.Hash, genesis.Serialize())
			if err != nil {
				return err
			}
			err = b.Put([]byte("l"), genesis.Hash)
			if err != nil {
				return err
			}
			tip = genesis.Hash
		} else {
			tip = b.Get([]byte("l"))
		}

		return nil
	})
	if err != nil {
		log.Fatalln("Failed to init blockchain from db: ", err)
	}

	bc := BlockChain{tip, db}
	return &bc
}

func (bc *BlockChain) AddBlock(data string) {
	var lastHash []byte

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		log.Fatalln("Failed to read lastHash in database: ", err)
	}

	newBlock := NewBlock(data, lastHash)

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err = b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			return err
		}
		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			return err
		}

		bc.tip = newBlock.Hash

		return nil
	})
	if err != nil {
		log.Fatalln("Failed to add NewBlock in database: ", err)
	}
}

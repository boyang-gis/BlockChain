package src

import (
	"fmt"
	"github.com/boltdb/bolt"
	"os"
)

const dbFile = "blockChain.db"
const blockBuket = "bucket"
const lastHashKey = "key"
const genesisInfo = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

type BlockChain struct {
	//blocks []*Block
	//Database operation handle, discarded

	db *bolt.DB
	//Tail, represents the hash of the last block.
	tail []byte
}

func isDBExist() bool {
	// If there is an error, it will be of type *PathError.
	_, err := os.Stat(dbFile)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

//create a blockchain database file
func InitBlockChain(address string) *BlockChain {
	if isDBExist() {
		fmt.Println("blockchain exist already, no need to create!")
		os.Exit(1)
	}

	//func Open(path string, mode os.FileMode, options *Options) (*DB, error)
	db, err := bolt.Open(dbFile, 0600, nil)
	CheckErr("InitBlockChain0", err)

	var lastHash []byte

	db.Update(func(tx *bolt.Tx) error {

		//Without a bucket, you have to create a creation block and return the data.
		coinbase := NewCoinBaseTx(address, genesisInfo)
		genesis := NewGenesisBlock(coinbase)
		bucket, err := tx.CreateBucket([]byte(blockBuket))
		CheckErr("InitBlockChain1", err)
		bucket.Put(genesis.Hash, genesis.Serialize())
		CheckErr("InitBlockChain2", err)
		bucket.Put([]byte(lastHashKey), genesis.Hash)
		CheckErr("InitBlockChain3", err)
		lastHash = genesis.Hash

		return nil
	})
	return &BlockChain{db, lastHash}
}

func GetBlockChainHandler() *BlockChain {
	if !isDBExist() {
		fmt.Println("Pls create blockchain first!")
		os.Exit(1)
	}

	//func Open(path string, mode os.FileMode, options *Options) (*DB, error)
	db, err := bolt.Open(dbFile, 0600, nil)
	CheckErr("GetBlockChainHandler1", err)

	var lastHash []byte

	//db.View()
	db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBuket))
		if bucket != nil {
			//Take the hash of the last block.
			lastHash = bucket.Get([]byte(lastHashKey))
		}

		return nil
	})
	return &BlockChain{db, lastHash}
}

func (bc *BlockChain) AddBlock(txs []*Transaction) {
	var prevBlockHash []byte

	bc.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBuket))
		if bucket == nil {
			os.Exit(1)
		}

		prevBlockHash = bucket.Get([]byte(lastHashKey))
		return nil
	})

	block := NewBlock(txs, prevBlockHash)
	err := bc.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBuket))
		if bucket == nil {
			os.Exit(1)
		}

		err := bucket.Put(block.Hash, block.Serialize())
		CheckErr("AddBlock1", err)
		err = bucket.Put([]byte(lastHashKey), block.Hash)
		CheckErr("AddBlock2", err)
		bc.tail = block.Hash
		return nil
	})
	CheckErr("AddBlock3", err)
}

//An iterator is an object, and it contains a cursor that moves forward (back) and completes the traversal of the certificate container.

type BlockChainIterator struct {
	currHash []byte
	db       *bolt.DB
}

//Create an iterator that is initialized to point to the last block.
func (bc *BlockChain) NewIterator() *BlockChainIterator {
	return &BlockChainIterator{currHash: bc.tail, db: bc.db}
}

func (it *BlockChainIterator) Next() (block *Block) {
	err := it.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBuket))
		if bucket == nil {
			return nil
		}

		data := bucket.Get(it.currHash)
		block = Deserialize(data)
		it.currHash = block.PreBlockHash
		return nil
	})
	CheckErr("Next()", err)
	return
}

//Returns a collection of utxo transactions that the specified address can control.
func (bc *BlockChain) FindUTXOTransactions(address string) []Transaction {
	//A collection of transactions containing the target utxo.
	var UTXOTransactions []Transaction

	//Store a collection of used utxo, map[transaction id] int64
	//ox1111111 : 0, 1 all transfered to Alice.
	spentUTXO := make(map[string][]int64)

	it := bc.NewIterator()
	for {
		//Traversing block
		block := it.Next()

		//Traversing transaction
		for _, tx := range block.Transactions {

			//Traversing input
			//Goal: Find the utxo that have been consumed and put them in a collection.
			//Two fields are needed to identify the used utxo: a. transaction ID, b. output's index.
			if !tx.IsCoinbase() {
				for _, input := range tx.TXInputs {
					if input.CanUnlockUTXOWith(address) {
						//map[txid][] int64
						//spentUTXO[string(tx.TXID)] = append(spentUTXO[string(tx.TXID)], input.Vout)
						spentUTXO[string(input.TXID)] = append(spentUTXO[string(input.TXID)], input.Vout)
					}
				}
			}

		OUTPUTS:
			//Traversing output
			//Goal: Find all the utxo you can control
			for currIndex, output := range tx.TXOutputs {
				//Check if the current output has been consumed. If it is consumed, then the next output check is performed.
				if spentUTXO[string(tx.TXID)] != nil {
					//Non-empty, representing the utxo consumed in the current transaction
					indexes := spentUTXO[string(tx.TXID)]
					for _, index := range indexes {
						//The current index and the consumed index comparison, if the same, indicate this output must be consumed. Skip directly and judge the next output
						if int64(currIndex) == index {
							continue OUTPUTS
						}
					}
				}

				//If the current address is the owner of this utxo, the candition is met
				if output.CanBeUnlockedWith(address) {
					UTXOTransactions = append(UTXOTransactions, *tx)
				}
			}
		}

		if len(block.PreBlockHash) == 0 {
			break
		}
	}
	return UTXOTransactions
}

//Find utxo that can be used at the specified address
func (bc *BlockChain) FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput
	txs := bc.FindUTXOTransactions(address)

	//Traversing transactions
	for _, tx := range txs {
		//Traversing output
		for _, utxo := range tx.TXOutputs {
			//utxo at the current address
			if utxo.CanBeUnlockedWith(address) {
				UTXOs = append(UTXOs, utxo)
			}
		}
	}
	return UTXOs
}

//validUTXOs/*A collection of reasonable utxo*/, total/*Return the sum of the amount of utxo*/ = bc.FindSuitableUTXOs(from, amount)
func (bc *BlockChain) FindSuitableUTXOs(address string, amount float64) (map[string][]int64, float64) {
	txs := bc.FindUTXOTransactions(address)
	validUTXOs := make(map[string][]int64)
	var total float64

FIND:
	//Traversing transactions
	for _, tx := range txs {
		outputs := tx.TXOutputs
		//Traversing outputs (utxo)
		for index, output := range outputs {
			if output.CanBeUnlockedWith(address) {
				//Determine whether the total amount of utxo currently collected is greater than the amount required.
				if total < amount {
					total += output.Value
					validUTXOs[string(tx.TXID)] = append(validUTXOs[string(tx.TXID)], int64(index))

				} else {
					break FIND
				}
			}
		}
	}
	return validUTXOs, total
}

package src

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"time"
)

type Block struct {
	/*
		Block Header
	*/
	// version
	Version int64
	// The hash of the previous block
	PreBlockHash []byte
	// The hash value of the current block, in order to simplify the code
	Hash []byte
	// Merkelgen
	MerKelRoot []byte
	// Timestamp
	TimeStamp int64
	// Difficulty value
	Bits int64
	// Random value
	Nonce int64

	//Transaction info
	//Data []byte
	Transactions []*Transaction
}

func (block *Block) Serialize() []byte {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(block)
	CheckErr("Serialize", err)
	return buffer.Bytes()
}

func Deserialize(data []byte) *Block {
	if len(data) == 0 {
		return nil
	}

	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&block)
	CheckErr("Deserialize", err)
	return &block
}

func NewBlock(txs []*Transaction, prevBlockHash []byte) *Block {
	var block Block
	block = Block{
		Version:      1,
		PreBlockHash: prevBlockHash,
		//Hash
		MerKelRoot:   []byte{},
		TimeStamp:    time.Now().Unix(),
		Bits:         targetBits,
		Nonce:        0,
		Transactions: txs}

	//block.SetHash()
	pow := NewProofOfWork(&block)
	nonce, hash := pow.Run()
	block.Nonce = nonce
	block.Hash = hash

	return &block
}

func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

// Roughly simulate the Merkel tree, splicing the hash values of the transaction to gernerate a root hash.
func (block *Block) HashTransactions() []byte {
	var txHashes [][]byte
	txs := block.Transactions

	//Traversing transaction
	for _, tx := range txs {
		//[]byte
		txHashes = append(txHashes, tx.TXID)
	}

	//Splicing two-dimensional slices to generate one-dimensional slices
	data := bytes.Join(txHashes, []byte{})
	hash /*[32]byte*/ := sha256.Sum256(data)
	return hash[:]
}

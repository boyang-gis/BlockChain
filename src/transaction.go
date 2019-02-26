package src

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"os"
)

const reward = 12.5

type Transaction struct {
	//transanction ID
	TXID []byte
	//input
	TXInputs []TXInput
	//output
	TXOutputs []TXOutput
}

type TXInput struct {
	//The transaction ID of the referenced output
	TXID []byte
	//The index value of the referenced output
	Vout int64
	//Unlock the script to indicate the conditions under which an output can be used
	ScriptSig string
}

//Check if the current user can unlock the referenced utxo
func (input *TXInput) CanUnlockUTXOWith(unlockData string) bool {
	return input.ScriptSig == unlockData
}

type TXOutput struct {
	//Amount paid to the payee
	Value float64
	//Lock the script to specify the address of the payee.
	ScriptPubKey string
}

//Check if the current user is the owner of this utxo.
func (output *TXOutput) CanBeUnlockedWith(unlockData string) bool {
	return output.ScriptPubKey == unlockData
}

//Set the transaction ID, which is a hash value.
func (tx *Transaction) SetTXID() {
	var buffer bytes.Buffer

	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(tx)

	CheckErr("SetTXID", err)
	hash := sha256.Sum256(buffer.Bytes())
	tx.TXID = hash[:]
}

//create a coinbase transaction, only the payee, no payer, is a reward transaction for completion
func NewCoinBaseTx(address string, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("reward to %s %d btc", address, reward)
	}

	input := TXInput{[]byte{}, -1, data}
	output := TXOutput{reward, address}

	tx := Transaction{[]byte{}, []TXInput{input}, []TXOutput{output}}
	tx.SetTXID()
	return &tx
}

func (tx *Transaction) IsCoinbase() bool {
	if len(tx.TXInputs) == 1 {
		if len(tx.TXInputs[0].TXID) == 0 && tx.TXInputs[0].Vout == -1 {
			return true
		}
	}
	return false
}

//Create a normal transaction, send function of send
func NewTransaction(from, to string, amount float64, bc *BlockChain) *Transaction {

	//map[string][] int64 key: transaction id, value: an indexed array that references output
	validUTXOs := make(map[string][]int64)
	var total float64
	validUTXOs /*A collection of reasonable utxo*/, total /*Return the sum of the amount of utxo*/ = bc.FindSuitableUTXOs(from, amount)

	//validUTXOs[0x11111111] = []int64{1}
	//validUTXOs[0x22222222] = []int64{0}
	//...
	//validUTXOs[0xnnnnnnnn] = []int64{n}
	if total < amount {
		fmt.Println("Not enough money!")
		os.Exit(1)
	}

	var inputs []TXInput
	var outputs []TXOutput

	//1.create inputs
	//Perform output to input conversion
	//Traversing the collection of valid utxo
	for txId, outputIndexes := range validUTXOs {
		for _, index := range outputIndexes {
			input := TXInput{[]byte(txId), int64(index), from}
			inputs = append(inputs, input)
		}
	}

	//2.create outputs
	//Pay to the other party
	output := TXOutput{amount, to}
	outputs = append(outputs, output)

	//Looking for change
	if total > amount {
		output := TXOutput{total - amount, from}
		outputs = append(outputs, output)
	}

	tx := Transaction{nil, inputs, outputs}
	tx.SetTXID()
	return &tx

}

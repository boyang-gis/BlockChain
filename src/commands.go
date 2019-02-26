package src

import "fmt"

func (cli *CLI) PrintChain() {
	bc := GetBlockChainHandler()
	defer bc.db.Close()

	//Print data
	it := bc.NewIterator()

	for {
		block := it.Next()

		fmt.Printf("version: %d\n", block.Version)
		fmt.Printf("PrevBlockHash: %x\n", block.PreBlockHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("TimeStamp: %d\n", block.TimeStamp)
		fmt.Printf("Bits: %d\n", block.Bits)
		fmt.Printf("Nonce: %d\n", block.Nonce)
		fmt.Printf("IsValid : %v\n", NewProofOfWork(block).IsValid())

		if len(block.PreBlockHash) == 0 {
			fmt.Println("print over!")
			break
		}
	}
}

func (cli *CLI) CreateChain(address string) {
	bc := InitBlockChain(address)
	defer bc.db.Close()
	fmt.Println("Create blockchain successfully!")
}

func (cli *CLI) GetBalance(address string) {
	bc := GetBlockChainHandler()
	defer bc.db.Close()

	utxos := bc.FindUTXO(address)

	//total amount
	var total float64 = 0

	//Traverse all utxo and get the total amount
	for _, utxo := range utxos {
		total += utxo.Value
	}

	fmt.Printf("The balance of %s is %f\n", address, total)
}

func (cli *CLI) Send(from, to string, amount float64) {
	bc := GetBlockChainHandler()
	defer bc.db.Close()

	tx := NewTransaction(from, to, amount, bc)
	bc.AddBlock([]*Transaction{tx})
	fmt.Println("send successfully!")
}

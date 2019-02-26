package src

import (
	"flag"
	"fmt"
	"os"
)

const usage = `
	createChain --address ADDRESS	"create a blockchain"
	send --from FROM --to TO --amount AMOUNT "send coin from FROM to TO"
	getBalance --address ADDRESS	"get balance of the address"
	printChain						"print all blocks"
`

const PrintChainCmdString = "printChain"
const CreateChainCmdString = "createChain"
const GetBalanceCmdString = "getBalance"
const SendCmdString = "send"

type CLI struct {
	//bc *BlockChain
}

func (cli *CLI) printUsage() {
	fmt.Println("Invalid input!")
	fmt.Println(usage)
	os.Exit(1)
}

func (cli *CLI) parameterCheck() {
	if len(os.Args) < 2 {
		cli.printUsage()
	}
}

func (cli *CLI) Run() {
	cli.parameterCheck()

	createChainCmd := flag.NewFlagSet(CreateChainCmdString, flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet(GetBalanceCmdString, flag.ExitOnError)
	printChainCmd := flag.NewFlagSet(PrintChainCmdString, flag.ExitOnError)
	sendCmd := flag.NewFlagSet(SendCmdString, flag.ExitOnError)

	//func (f *FlagSet) String(name string, value string, usage string) *string
	//create blockchain related parameters
	createChainCmdPara := createChainCmd.String("address", "", "address info!")

	//balance related parameters
	getBalanceCmdPara := getBalanceCmd.String("address", "", "address info!")

	//send related parameters
	fromPara := sendCmd.String("from", "", "send address info!")
	toPara := sendCmd.String("to", "", "to address info!")
	amountPara := sendCmd.Float64("amount", 0, "amount info!")

	switch os.Args[1] {
	case CreateChainCmdString:
		//create blockchain
		err := createChainCmd.Parse(os.Args[2:])
		CheckErr("Run0()", err)
		if createChainCmd.Parsed() {
			if *createChainCmdPara == "" {
				fmt.Println("address should not be empty!")
				cli.printUsage()
			}

			cli.CreateChain(*createChainCmdPara)
		}

	case SendCmdString:
		//send transaction
		err := sendCmd.Parse(os.Args[2:])
		CheckErr("Run1()", err)
		if sendCmd.Parsed() {
			if *fromPara == "" || *toPara == "" || *amountPara <= 0 {
				fmt.Println("send cmd parameters invalid!")
				cli.printUsage()
			}

			cli.Send(*fromPara, *toPara, *amountPara)
		}

	case GetBalanceCmdString:
		//get balance
		err := getBalanceCmd.Parse(os.Args[2:])
		CheckErr("Run2()", err)
		if getBalanceCmd.Parsed() {
			if *getBalanceCmdPara == "" {
				fmt.Println("address should not be empty!")
				cli.printUsage()
			}

			cli.GetBalance(*getBalanceCmdPara)
		}

	case PrintChainCmdString:
		//print
		err := printChainCmd.Parse(os.Args[2:])
		CheckErr("Run3()", err)
		if printChainCmd.Parsed() {
			cli.PrintChain()
		}

	default:
		cli.printUsage()

	}
}

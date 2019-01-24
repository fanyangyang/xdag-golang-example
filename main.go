package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

/**
a wallet block, if you want get it all transaction, ask fro /api/block/{wallet block address}
and then iterator all block as address, that all are wallet block's transaction
the block as address has two direction, which one is input, specify that is a transaction the wallet transfer in coins, ,
relatively the one whose direction is output specify that a transaction the wallet transfer out coins.

 */

const API_URL = "https://explorer.xdag.io"

const API_BLOCK = "/api/block/"
const API_STATUS = "/api/status"

type JsonBlock struct {
	Time               string              `json:"time"`
	Timestamp          string              `json:"timestamp"`
	Flags              string              `json:"flags"`
	State              string              `json:"state"`
	FilePos            string              `json:"file_pos"`
	Hash               string              `json:"hash"`
	Remark             string              `json:"remark"`
	Difficulty         string              `json:"difficulty"`
	BalanceAddress     string              `json:"balance_address"`
	Balance            string              `json:"balance"`
	BlockAsTransaction []JsonBlockInternal `json:"block_as_transaction"`
	BlockAsAddress     []JsonBlockInternal `json:"block_as_address"`
	TotalEarnings      float64             `json:"total_earnings"`
	TotalSpendings     float64             `json:"total_spendings"`
	Kind               string              `json:"kind"` // "Main block
}

type JsonBlockInternal struct {
	Address    string `json:"address"`
	CreateTime string `json:"time"`
	Direction  string `json:"direction"`
	Amount     string `json:"amount"`
	Remark     string `json:"remark"`
}

func main() {
	userPersonalAddress := ""
	userExchangeAddress := ""
	var lastCheckTime int64 = 0
	for {
		fmt.Println(CheckUserBalance(userExchangeAddress, userPersonalAddress, lastCheckTime))

	}

}

func GetStatus() (ok bool) {
	request, err := http.NewRequest("GET", API_URL+API_STATUS, nil)
	if err != nil {
		fmt.Println(err)
		return false
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		fmt.Println(err)
		return false
	}

	defer response.Body.Close()

	body, _ := ioutil.ReadAll(response.Body)

	fmt.Println("respbody : ", string(body))
	return ok
}

//# Check user charge amount
func CheckUserBalance(exchangeAddress, personalAddress string, lastCheckTime int64) (balance float64) {
	transactions := GetTransaction(exchangeAddress)
	return CheckBalanceInTransaction(transactions, exchangeAddress, personalAddress, lastCheckTime)
}

//# Check the charge sum of specified address in transactions
func CheckBalanceInTransaction(transactions []JsonBlockInternal, exchangeAddress, personalAddress string, fromTime int64) (balance float64) {
	for _, tx := range transactions {
		// tx.Address here can specify a unique transaction, so you can use tx.Address as db's key for skip getBlockInfo repeatedly
		// mark exchangeAddress's all deposit transaction
		if tx.Direction == "input" {
			createTime, err := time.Parse("2006-01-02 15:04:05.000", tx.CreateTime)
			if err != nil {
				continue
			}

			if fromTime > createTime.Unix() {
				continue
			}

			jsonBlock := GetBlockInfo(tx.Address)
			if jsonBlock.State != "Accepted" {
				fmt.Println("found a transaction status is still not accepted, please pay attention")
				continue
			}
			if jsonBlock.BlockAsAddress != nil && len(jsonBlock.BlockAsTransaction) > 0 {
				for _, field := range jsonBlock.BlockAsTransaction {
					if field.Direction == "input" && field.Address == personalAddress {
						fmt.Printf("found exchangeAddress = %s's deposit, ready to check this transaction exist or not in db\n", exchangeAddress)
						amount, _ := strconv.ParseFloat(field.Amount, 64)
						balance += amount
					}
				}
			}
		}

		// mark exchangeAddress's all withdraw transaction
		if tx.Direction == "output" {
			createTime, err := time.Parse("2006-01-02 15:04:05.000", tx.CreateTime)
			if err != nil {
				continue
			}
			if fromTime > createTime.Unix() {
				continue
			}

			jsonBlock := GetBlockInfo(tx.Address)
			if jsonBlock.State != "Accepted" {
				fmt.Println("found a transaction status is still not accepted, please pay attention")
				continue
			}
			if jsonBlock.BlockAsTransaction != nil && len(jsonBlock.BlockAsTransaction) > 0 {
				for _, field := range jsonBlock.BlockAsTransaction {
					if field.Direction == "input" && field.Address == exchangeAddress {
						fmt.Printf("found exchangeAddress = %s's withdraw, ready to check this transaction exist or not in db\n", exchangeAddress)
						amount, _ := strconv.ParseFloat(field.Amount, 64)
						balance -= amount
					}
				}
			}
		}

	}
	return
}

func GetTransaction(adderess string) (transactions []JsonBlockInternal) {
	jsonBlock := SendRequest(API_URL + API_BLOCK + adderess)
	return jsonBlock.BlockAsAddress
}

func GetBlockInfo(blockAddress string) (jsonBlock JsonBlock) {
	return SendRequest(API_URL + API_BLOCK + blockAddress)
	//return jsonBlock.BlockAsTransaction // check jsonBlock's BlockAsTransaction
}

func SendRequest(api string) (jsonBlock JsonBlock) {
	request, err := http.NewRequest("GET", api, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		fmt.Println(err)
		fmt.Println("try 3 times start")
		for i := 0; i < 3; i++ {
			fmt.Println("sleep 2 sec")
			time.Sleep(time.Second * 2)
			response, err = http.DefaultClient.Do(request)
			if err == nil && response.Status == "200 OK" {
				fmt.Println("got right response continue")
				break
			}
		}
	}
	if err != nil || response.Status != "200 OK" {
		fmt.Println("response status : ", response.Status, "and error : ", err)
		return
	}

	defer response.Body.Close()

	body, _ := ioutil.ReadAll(response.Body)

	fmt.Println("respbody : ", string(body))

	err = json.Unmarshal(body, &jsonBlock)
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

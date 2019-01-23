package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

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
	Address        string `json:"address"`
	CreateTime     string `json:"time"`
	Direction      string `json:"direction"`
	Amount         string `json:"amount"`
	Remark         string `json:"remark"`
}

func main() {
	address := ""   //TODO you wallet address paste here
	var lastCheckTime int64 = 0

	fmt.Println(CheckUserBalance(address, lastCheckTime))

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

func GetTransaction(adderess string) (transactions []JsonBlockInternal) {
	jsonBlock := SendRequest(API_URL + API_BLOCK + adderess)
	return jsonBlock.BlockAsAddress
}

func GetBlockInfo(blockAddress string) (blockInfo []JsonBlockInternal) {
	jsonBlock := SendRequest(API_URL + API_BLOCK + blockAddress)
	return jsonBlock.BlockAsTransaction
}

//# Check the charge sum of specified address in transactions
func CheckBalanceInTransaction(transactions []JsonBlockInternal, address string, fromTime int64) (balance float64) {
	for _, tx := range transactions {
		if tx.Direction == "input" {
			createTime, err := time.Parse("2006-01-02 15:04:05.000", tx.CreateTime)
			if err != nil {
				continue
			}

			if fromTime > createTime.Unix() {
				continue
			}

			blockFields := GetBlockInfo(tx.Address)
			if blockFields != nil && len(blockFields) > 0{
				for _, field := range blockFields {
					if field.Direction == "output" && field.Address == address {
						amount, _ := strconv.ParseFloat(field.Amount, 64)
						balance += amount
					}
				}
			}
		}
	}
	return
}

//# Check user charge amount
func CheckUserBalance(address string, lastCheckTime int64) (balance float64) {
	transactions := GetTransaction(address)
	return CheckBalanceInTransaction(transactions, address, lastCheckTime)
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

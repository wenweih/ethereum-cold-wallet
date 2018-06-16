package main

import (
	"encoding/json"
	"errors"
	"math/big"
	"strconv"
	"strings"

	"github.com/parnurzeal/gorequest"
)

var (
	// APIKEY ethereum oauth token
	APIKEY  string
	request *gorequest.SuperAgent
)

// AccountRespBody EtherScan Response body
type AccountRespBody struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  string `json:"result"`
}

// ProxyRespBody EtherScan Response body
type ProxyRespBody struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  string `json:"result"`
}

func (es EtherScan) getBalance(address string) (*big.Int, error) {
	resp, body, err := request.Get(etherscan.URL).Query(map[string]interface{}{
		"module":  "account",
		"action":  "balance",
		"address": address,
		"tag":     "latest",
		"apikey":  APIKEY,
	}).End()

	if err != nil {
		return nil, errors.New(strings.Join([]string{"etherscan: get balance error:", address, err[0].Error()}, " "))
	}
	if handleStatus(resp) {
		var (
			respBody = new(AccountRespBody)
			balance  = new(big.Int)
		)
		if err := json.Unmarshal([]byte(body), respBody); err != nil {
			return nil, errors.New("etherscan getBalance Unmarshal error")
		}
		balance.SetString(respBody.Result, 10)
		return balance, nil
	}
	return nil, errors.New("etherscan get balance error")
}

func (es EtherScan) getGasPrice() (*big.Int, error) {
	resp, body, err := request.Get(etherscan.URL).Query(map[string]interface{}{
		"module": "proxy",
		"action": "eth_gasPrice",
		"apikey": APIKEY,
	}).End()

	if err != nil {
		return nil, errors.New(strings.Join([]string{"etherscan: get balance error:", err[0].Error()}, " "))
	}
	if handleStatus(resp) {
		var (
			respBody = new(ProxyRespBody)
			gasPrice = new(big.Int)
		)
		if err := json.Unmarshal([]byte(body), respBody); err != nil {
			return nil, errors.New("etherscan getBalance Unmarshal error")
		}
		resultWithoutHex := strings.Replace(respBody.Result, "0x", "", -1)
		gasPrice.SetString(resultWithoutHex, 10)
		return gasPrice, nil
	}
	return nil, errors.New("etherscan get gasPrice error")
}

func (es EtherScan) getAccountNonce(address string) (*uint64, error) {
	resp, body, err := request.Get(etherscan.URL).Query(map[string]interface{}{
		"module":  "proxy",
		"action":  "eth_getTransactionCount",
		"address": address,
		"tag":     "latest",
		"apikey":  APIKEY,
	}).End()

	if err != nil {
		return nil, errors.New(strings.Join([]string{"etherscan: get account nonce error:", address, err[0].Error()}, " "))
	}
	if handleStatus(resp) {
		var respBody = new(ProxyRespBody)
		if err := json.Unmarshal([]byte(body), respBody); err != nil {
			return nil, errors.New("etherscan getBalance Unmarshal error")
		}

		nonce, _ := strconv.ParseUint(strings.Replace(respBody.Result, "0x", "", -1), 16, 64)
		return &nonce, nil
	}
	return nil, errors.New("etherscan get account nonce error")
}

func (es EtherScan) etherscanConstructTxField(address string) (*big.Int, *uint64, *big.Int, error) {
	balance, err := es.getBalance(address)
	if err != nil {
		return nil, nil, nil, err
	}

	err = balanceIsLessThanConfig(address, balance)
	if err != nil {
		return nil, nil, nil, err
	}

	accountNonce, err := es.getAccountNonce(address)
	if err != nil {
		return nil, nil, nil, errors.New(strings.Join([]string{"etherscan: get account nonce error:", address, err.Error()}, " "))
	}

	gasPrice, err := es.getGasPrice()
	if err != nil {
		return nil, nil, nil, err
	}

	return balance, accountNonce, gasPrice, nil
}

func handleStatus(resp gorequest.Response) bool {
	if resp.StatusCode == 200 {
		return true
	}
	return false
}

func init() {
	APIKEY = etherscan.Key
	request = gorequest.New()
}

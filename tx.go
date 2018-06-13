package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/big"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
)

// Tx 交易结构体
type Tx struct {
	From  string  `json:"from"`
	To    string  `json:"to"`
	TxHex string  `json:"txhex"`
	Value big.Int `json:"value"`
	Nonce uint64  `json:"nonce"`
}

func exportHexTx(from, to, txHex *string, value *big.Int, nonce *uint64, signed bool) error {
	tx := &Tx{
		From:  *from,
		To:    *to,
		TxHex: *txHex,
		Value: *value,
		Nonce: *nonce,
	}

	bTx, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	var configurePath string
	if signed {
		configurePath = config.SignedTx
	} else {
		configurePath = config.RawTx
	}
	TxPath, err := mkdirBySlice([]string{HomeDir(), configurePath})
	if err != nil {
		return errors.New(strings.Join([]string{"Could not create directory", err.Error()}, " "))
	}

	txFileName := strings.Join([]string{"from", *from, "json"}, ".")
	txfile := strings.Join([]string{*TxPath, txFileName}, "/")
	if err := ioutil.WriteFile(txfile, bTx, 0600); err != nil {
		return errors.New(strings.Join([]string{"Failed to write keyfile to", err.Error()}, " "))
	}
	log.Infoln("Exported HexTx to", txfile)
	return nil
}

func constructTx(nodeClient *ethclient.Client, nonce uint64, balance *big.Int, hexAddressFrom, hexAddressTo *string) (*string, *string, *string, *big.Int, error) {
	gasLimit := uint64(21000) // in units
	gasPrice, err := nodeClient.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, nil, nil, nil, errors.New(strings.Join([]string{"get gasPrice error", err.Error()}, " "))
	}

	if !common.IsHexAddress(*hexAddressTo) {
		return nil, nil, nil, nil, errors.New(strings.Join([]string{*hexAddressTo, "invalidate"}, " "))
	}
	var (
		txFee = new(big.Int)
		value = new(big.Int)
	)

	txFee = txFee.Mul(gasPrice, big.NewInt(int64(gasLimit)))
	value = value.Sub(balance, txFee)

	tx := types.NewTransaction(nonce, common.HexToAddress(*hexAddressTo), value, gasLimit, gasPrice, nil)
	rawTxHex, err := encodeTx(tx)
	if err != nil {
		return nil, nil, nil, nil, errors.New(strings.Join([]string{"encode raw tx error", err.Error()}, " "))
	}
	return hexAddressFrom, hexAddressTo, rawTxHex, value, nil
}

func decodeTx(txHex *string) (*types.Transaction, error) {
	txc, err := hexutil.Decode(*txHex)
	if err != nil {
		return nil, err
	}

	var txde types.Transaction

	t, err := &txde, rlp.Decode(bytes.NewReader(txc), &txde)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func encodeTx(tx *types.Transaction) (*string, error) {
	txb, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return nil, err
	}
	txHex := hexutil.Encode(txb)
	return &txHex, nil
}

func signTxCmd() {
	files, err := ioutil.ReadDir(strings.Join([]string{HomeDir(), config.RawTx}, "/"))
	if err != nil {
		log.Fatalln("read raw tx error", err.Error())
	}

	for _, file := range files {
		fileName := file.Name()
		tx, err := readTxHex(&fileName, false)
		if err != nil {
			log.Errorln(err.Error())
		}

		rawtxHex := tx.TxHex
		fromHex := tx.From

		from, to, signedTxHex, value, nonce, err := signTx(&rawtxHex, &fromHex)
		if err != nil {
			log.Errorln(strings.Join([]string{"sign tx from", *from, "error", err.Error()}, " "))
			continue
		}
		if err := exportHexTx(from, to, signedTxHex, value, nonce, true); err != nil {
			log.Errorln(strings.Join([]string{"export signed tx hex to", fileName, "error, issue by address:", *from, err.Error()}, " "))
			continue
		}
	}
}

func sendTxCmd(nodeClient *ethclient.Client) {
	files, err := ioutil.ReadDir(strings.Join([]string{HomeDir(), config.SignedTx}, "/"))
	if err != nil {
		log.Fatalln("read raw tx error", err.Error())
	}

	for _, file := range files {
		fileName := file.Name()
		tx, err := readTxHex(&fileName, true)
		if err != nil {
			log.Errorln(err.Error())
		}

		signedTxHex := tx.TxHex
		to := config.To
		hash, err := sendTx(&signedTxHex, &to, nodeClient)
		if err != nil {
			log.Errorln("send tx: ", fileName, "fail", err.Error())
		} else {
			log.Infoln("send tx: ", *hash, "success")
		}
	}
}

func readTxHex(fileName *string, signed bool) (*Tx, error) {
	var filePath string
	if signed {
		filePath = strings.Join([]string{HomeDir(), config.SignedTx, *fileName}, "/")
	} else {
		filePath = strings.Join([]string{HomeDir(), config.RawTx, *fileName}, "/")
	}

	bRawTx, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, errors.New(strings.Join([]string{"can't read", filePath, err.Error()}, " "))
	}

	var tx Tx
	if err := json.Unmarshal(bRawTx, &tx); err != nil {
		return nil, errors.New(strings.Join([]string{"can't Unmarshal", filePath, "to RawTx struct"}, " "))
	}
	return &tx, nil
}

func signTx(txHex, fromAddressHex *string) (*string, *string, *string, *big.Int, *uint64, error) {
	tx, err := decodeTx(txHex)
	if err != nil {
		return nil, nil, nil, nil, nil, errors.New(strings.Join([]string{"decode tx error", err.Error()}, " "))
	}

	key, err := decodeKS2Key(fromAddressHex)
	if err != nil {
		return nil, nil, nil, nil, nil, errors.New(strings.Join([]string{"decode keystore to key error", err.Error()}, " "))
	}

	// https://github.com/ethereum/EIPs/blob/master/EIPS/eip-155.md
	// chain id
	// 1 Ethereum mainnet
	// 61 Ethereum Classic mainnet
	// 62 Ethereum Classic testnet
	// 1337 Geth private chains (default)
	var chainID *big.Int
	switch config.NetMode {
	case "private":
		chainID = big.NewInt(1337)
	case "mainnet":
		chainID = big.NewInt(1)
	default:
		return nil, nil, nil, nil, nil, errors.New("you must set net_mode in configure")
	}
	signtx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), key.PrivateKey)
	if err != nil {
		return nil, nil, nil, nil, nil, errors.New(strings.Join([]string{"sign tx error", err.Error()}, " "))
	}
	msg, err := signtx.AsMessage(types.NewEIP155Signer(chainID))
	if err != nil {
		return nil, nil, nil, nil, nil, errors.New(strings.Join([]string{"tx to msg error", err.Error()}, " "))
	}

	from := msg.From().Hex()
	to := msg.To().Hex()
	value := msg.Value()
	nonce := msg.Nonce()
	signTxHex, err := encodeTx(signtx)
	return &from, &to, signTxHex, value, &nonce, nil
}

func sendTx(signTxHex, to *string, nodeClient *ethclient.Client) (*string, error) {
	signTx, _ := decodeTx(signTxHex)
	if strings.Compare(strings.ToLower(signTx.To().Hex()), *to) != 0 {
		return nil, errors.New("decode tx and to field error")
	}

	if err := nodeClient.SendTransaction(context.Background(), signTx); err != nil {
		return nil, err
	}
	h := signTx.Hash().Hex()
	return &h, nil
}

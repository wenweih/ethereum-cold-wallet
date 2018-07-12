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
	Hash  string  `json:"hash"`
}

func exportHexTx(from, to, txHex, hash string, value *big.Int, nonce *uint64, signed bool) error {
	tx := &Tx{
		From:  from,
		To:    to,
		TxHex: txHex,
		Value: *value,
		Nonce: *nonce,
		Hash:  hash,
	}

	bTx, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	var configurePath, txFileName string
	if signed {
		configurePath = config.SignedTx
		txFileName = strings.Join([]string{"signed_from", from, "json"}, ".")
	} else {
		configurePath = config.RawTx
		txFileName = strings.Join([]string{"unsign_from", from, "json"}, ".")
	}
	TxPath, err := mkdirBySlice([]string{HomeDir(), configurePath})
	if err != nil {
		return errors.New(strings.Join([]string{"Could not create directory", err.Error()}, " "))
	}

	txfile := strings.Join([]string{*TxPath, txFileName}, "/")
	if err := ioutil.WriteFile(txfile, bTx, 0600); err != nil {
		return errors.New(strings.Join([]string{"Failed to write tx to", err.Error()}, " "))
	}
	log.Infoln("Exported HexTx to", txfile)
	return nil
}

func constructTxCmd() {
	ormDB := ormBbAlias{dbConn()}
	ormDB.DBMigrate()
	defer ormDB.Close()
	ormDB.csv2db()

	var subAddresses []*SubAddress
	ormDB.Find(&subAddresses)
	for _, subaddress := range subAddresses {
		from, balance, pendingNonceAt, gasPrice, err := ormDB.constructTxField(subaddress.Address)
		if err != nil {
			log.Warnln(err.Error())
			continue
		}

		to := randomPickFromSlice(config.To)
		if err := applyWithdrawAndConstructRawTx(balance, gasPrice, pendingNonceAt, *from, to); err != nil {
			log.Warnln(err.Error())
		}
	}
}

func applyWithdrawAndConstructRawTx(balance, gasPrice *big.Int, nonce *uint64, from, to string) error {
	if err := balanceIsLessThanConfig(from, balance); err != nil {
		return err
	}

	fromHex, toHex, rawTxHex, txHashHex, value, err := constructTx(*nonce, balance, gasPrice, from, to)
	if err != nil {
		return errors.New(strings.Join([]string{"constructTx error", err.Error()}, " "))
	}
	if err := exportHexTx(*fromHex, *toHex, *rawTxHex, *txHashHex, value, nonce, false); err != nil {
		return errors.New(strings.Join([]string{"sub address:", from, "hased applied withdraw, but fail to export rawTxHex to ", config.RawTx, err.Error()}, " "))
	}
	return nil
}

func constructTx(nonce uint64, balance, gasPrice *big.Int, hexAddressFrom, hexAddressTo string) (*string, *string, *string, *string, *big.Int, error) {
	gasLimit := uint64(21000) // in units

	if !common.IsHexAddress(hexAddressTo) {
		return nil, nil, nil, nil, nil, errors.New(strings.Join([]string{hexAddressTo, "invalidate"}, " "))
	}

	var (
		txFee = new(big.Int)
		value = new(big.Int)
	)

	txFee = txFee.Mul(gasPrice, big.NewInt(int64(gasLimit)))
	value = value.Sub(balance, txFee)

	tx := types.NewTransaction(nonce, common.HexToAddress(hexAddressTo), value, gasLimit, gasPrice, nil)
	rawTxHex, err := encodeTx(tx)
	if err != nil {
		return nil, nil, nil, nil, nil, errors.New(strings.Join([]string{"encode raw tx error", err.Error()}, " "))
	}
	txHashHex := tx.Hash().Hex()
	return &hexAddressFrom, &hexAddressTo, rawTxHex, &txHashHex, value, nil
}

func nodeConstructTxField(node, address string) (*big.Int, *uint64, *big.Int, error) {
	client, err := nodeClient(node)
	if err != nil {
		return nil, nil, nil, err
	}
	balance, nonce, gasPrice, err := getBalanceAndPendingNonceAtAndGasPrice(client, address)
	if err != nil {
		return nil, nil, nil, err
	}
	return balance, nonce, gasPrice, nil
}

func getBalanceAndPendingNonceAtAndGasPrice(node *ethclient.Client, address string) (*big.Int, *uint64, *big.Int, error) {
	ctx := context.Background()
	balance, err := node.BalanceAt(ctx, common.HexToAddress(address), nil)
	if err != nil {
		return nil, nil, nil, errors.New(strings.Join([]string{"Failed to get ethereum balance from address:", address, err.Error()}, " "))
	}

	if err := balanceIsLessThanConfig(address, balance); err != nil {
		return nil, nil, nil, err
	}

	pendingNonceAt, err := node.PendingNonceAt(ctx, common.HexToAddress(address))
	if err != nil {
		return nil, nil, nil, errors.New(strings.Join([]string{"Failed to get account nonce from address:", address, err.Error()}, " "))
	}

	gasPrice, err := node.SuggestGasPrice(ctx)
	if err != nil {
		return nil, nil, nil, errors.New(strings.Join([]string{"get gasPrice error", err.Error()}, " "))
	}

	return balance, &pendingNonceAt, gasPrice, nil

}

func decodeTx(txHex string) (*types.Transaction, error) {
	txc, err := hexutil.Decode(txHex)
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
			continue
		}

		from, to, signedTxHex, hash, value, nonce, err := signTx(tx)
		if err != nil {
			log.Errorln(strings.Join([]string{"sign tx from", tx.From, "error", err.Error()}, " "))
			continue
		}
		if err := exportHexTx(*from, *to, *signedTxHex, *hash, value, nonce, true); err != nil {
			log.Errorln(strings.Join([]string{"export signed tx hex to", fileName, "error, issue by address:", *from, err.Error()}, " "))
			continue
		}
	}
}

func signTx(simpletx *Tx) (*string, *string, *string, *string, *big.Int, *uint64, error) {
	txHex := simpletx.TxHex
	fromAddressHex := simpletx.From
	tx, err := decodeTx(txHex)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, errors.New(strings.Join([]string{"decode tx error", err.Error()}, " "))
	}

	if Contains(config.To, tx.To().Hex()) {
		log.Infoln("签名交易：", tx.Hash().Hex(), " To:", tx.To().Hex())
	} else {
		promptSign(tx.To().Hex())
	}

	key, err := decodeKS2Key(fromAddressHex)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, errors.New(strings.Join([]string{"decode keystore to key error:", err.Error()}, " "))
	}

	// https://github.com/ethereum/EIPs/blob/master/EIPS/eip-155.md
	// chain id
	// 1 Ethereum mainnet
	// 61 Ethereum Classic mainnet
	// 62 Ethereum Classic testnet
	// 1337 Geth private chains (default)
	var chainID *big.Int
	switch config.NetMode {
	case "privatenet":
		chainID = big.NewInt(1337)
	case "mainnet":
		chainID = big.NewInt(1)
	default:
		return nil, nil, nil, nil, nil, nil, errors.New("you must set net_mode in configure")
	}
	signtx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), key.PrivateKey)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, errors.New(strings.Join([]string{"sign tx error", err.Error()}, " "))
	}
	msg, err := signtx.AsMessage(types.NewEIP155Signer(chainID))
	if err != nil {
		return nil, nil, nil, nil, nil, nil, errors.New(strings.Join([]string{"tx to msg error", err.Error()}, " "))
	}

	from := msg.From().Hex()
	to := msg.To().Hex()
	value := msg.Value()
	nonce := msg.Nonce()
	signTxHex, err := encodeTx(signtx)
	hash := signtx.Hash().Hex()
	return &from, &to, signTxHex, &hash, value, &nonce, nil
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
		hash, err := sendTx(signedTxHex, nodeClient)
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

func sendTx(signTxHex string, nodeClient *ethclient.Client) (*string, error) {
	signTx, err := decodeTx(signTxHex)
	if err != nil {
		return nil, errors.New(strings.Join([]string{"Send tx error:", "decode tx error", err.Error()}, " "))
	}

	if !Contains(config.To, signTx.To().Hex()) {
		return nil, errors.New(strings.Join([]string{"Send tx error: ", signTx.To().Hex(), "is not contained in configure to value"}, " "))
	}

	if err := nodeClient.SendTransaction(context.Background(), signTx); err != nil {
		return nil, err
	}
	h := signTx.Hash().Hex()
	return &h, nil
}

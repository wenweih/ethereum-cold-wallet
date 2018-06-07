package main

import (
	"bytes"
	"context"
	"errors"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
)

func exportRawHexTx() {

}
func constructTx(nodeClient *ethclient.Client, nonce uint64, balance *big.Int, hexAddressFrom, hexAddressTo string) (*common.Address, *common.Address, *string, *big.Int, error) {
	gasLimit := uint64(21000) // in units
	gasPrice, err := nodeClient.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, nil, nil, nil, errors.New(strings.Join([]string{"get gasPrice error", err.Error()}, " "))
	}

	if !common.IsHexAddress(hexAddressTo) {
		return nil, nil, nil, nil, errors.New(strings.Join([]string{hexAddressTo, "invalidate"}, " "))
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
		return nil, nil, nil, nil, err
	}

	from := common.HexToAddress(hexAddressFrom)
	to := common.HexToAddress(hexAddressTo)
	return &from, &to, rawTxHex, value, nil
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

func signTxs() {

}

func signTx(txHex, fromAddressHex *string) (*string, error) {
	tx, err := decodeTx(txHex)
	if err != nil {
		return nil, errors.New(strings.Join([]string{"decode tx error", err.Error()}, " "))
	}

	key, err := decodeKS2Key(fromAddressHex)
	if err != nil {
		return nil, errors.New(strings.Join([]string{"decode keystore to key error", err.Error()}, " "))
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
		return nil, errors.New("you must set net_mode in configure")
	}
	signtx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), key.PrivateKey)
	if err != nil {
		return nil, errors.New(strings.Join([]string{"sign tx error", err.Error()}, " "))
	}
	return encodeTx(signtx)
}

func sendTx(signTxHex, to *string, nodeClient *ethclient.Client) (*string, error) {
	signTx, _ := decodeTx(signTxHex)
	if strings.ToLower(signTx.To().Hex()) != config.To {
		return nil, errors.New("decode tx and to file error")
	}

	if err := nodeClient.SendTransaction(context.Background(), signTx); err != nil {
		return nil, err
	}
	h := signTx.Hash().Hex()
	return &h, nil
}

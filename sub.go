package main

import (
	"context"
	"errors"
	"math/big"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func subNewBlockCmd() {
	ormDB := ormBbAlias{dbConn()}
	ormDB.DBMigrate()
	defer ormDB.Close()

	ctx := context.Background()
	nodeClient, err := ethclient.Dial(config.EthRPC)
	if err != nil {
		log.Fatalln(err.Error())
	}

	blockCh := make(chan *types.Header)
	sub, err := nodeClient.SubscribeNewHead(ctx, blockCh)

	if err != nil {
		log.Error(err.Error())
	}

	var (
		// maintain orderHeight and increase 1 each subscribe callback, because head.number would jump blocks
		orderHeight = new(big.Int)
	)
	for {
		select {
		case err := <-sub.Err():
			log.Fatalln(err.Error())
		case head := <-blockCh:
			ordertmp, err := subHandle(orderHeight, head, nodeClient)
			if err != nil {
				log.Errorln(err.Error())
			}
			orderHeight = ordertmp
		}
	}
}

func subHandle(orderHeight *big.Int, head *types.Header, nodeClient *ethclient.Client) (*big.Int, error) {
	ctx := context.Background()
	number := head.Number
	originBlock, err := nodeClient.BlockByNumber(ctx, number)
	if err != nil {
		return nil, errors.New(strings.Join([]string{"get origin block error, height:", number.String(), err.Error()}, " "))
	}

	number.Sub(number, big.NewInt(1))
	parentBlock, err := nodeClient.BlockByNumber(ctx, number)
	if err != nil {
		return nil, errors.New(strings.Join([]string{"get parent block error", number.String(), err.Error()}, " "))
	}

	if originBlock.ParentHash().Hex() != parentBlock.Hash().Hex() {
		return nil, errors.New(strings.Join([]string{"uncle block, stable block's  height:", originBlock.Number().String()}, " "))
	}

	if orderHeight.Cmp(big.NewInt(0)) == 0 {
		orderHeight = originBlock.Number()
	}

	var pushJumpBlock string
	log.Infoln("sub message coming,", "order height:", orderHeight.Int64(), "sub block height:", originBlock.Number().Int64())
	for blockNumber := orderHeight.Int64(); blockNumber <= originBlock.Number().Int64(); blockNumber++ {
		block, err := nodeClient.BlockByNumber(ctx, big.NewInt(blockNumber))
		if err != nil {
			log.Warnln("Get block error, height:", blockNumber)
			continue
		}

		if blockNumber < originBlock.Number().Int64() {
			pushJumpBlock = "jump"
		} else {
			pushJumpBlock = ""
		}
		log.Infoln("New", pushJumpBlock, "block, Height:", block.Number().String(), "blockHash:", block.Hash().Hex())
		iteratorBlockTx(block, nodeClient)
		orderHeight.Add(orderHeight, big.NewInt(1))
	}
	return orderHeight, nil
}

func iteratorBlockTx(block *types.Block, nodeClient *ethclient.Client) {
	txs := block.Transactions()
	for _, tx := range txs {
		var to string
		pto := tx.To()

		// contract creation transaction, to field is empty
		if pto != nil {
			to = (*pto).Hex()
		} else {
			continue
		}
		log.Infoln(to)
	}
}

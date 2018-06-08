package main

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"strings"

	"github.com/olivere/elastic"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func subNewBlockCmd() {
	ctx := context.Background()
	esClient, err := elastic.NewClient(elastic.SetURL(config.ElasticURL), elastic.SetSniff(config.ElasticSniff))
	if err != nil {
		log.Fatalln(err.Error())
	}
	esClient.DeleteIndex("eth_sub_address").Do(ctx)
	csv2es(ctx, esClient)
	nodeClient, err := ethclient.Dial(config.EthRPC)
	if err != nil {
		log.Fatalln(err.Error())
	}

	blockCh := make(chan *types.Header)
	sub, err := nodeClient.SubscribeNewHead(ctx, blockCh)
	if err != nil {
		log.Error(err.Error())
	}

	for {
		select {
		case err := <-sub.Err():
			log.Fatalln(err.Error())
		case head := <-blockCh:
			block, err := nodeClient.BlockByNumber(ctx, head.Number)
			if err != nil {
				log.Fatalln(err.Error())
			}
			txs := block.Transactions()
			for _, tx := range txs {
				to := tx.To().Hex()
				from, balance, pendingNonceAt, err := addressWithAmount(nodeClient, esClient, to)
				if err != nil {
					log.Warnln(err.Error())
					continue
				}

				if err := applyWithdrawAndConstructRawTx(balance, pendingNonceAt, nodeClient, from, &(config.To)); err != nil {
					log.Errorln(err.Error())
				}
			}
		}
	}
}

func applyWithdrawAndConstructRawTx(balance *big.Int, nonce *uint64, client *ethclient.Client, from, to *string) error {
	balanceDecimal, _ := decimal.NewFromString(balance.String())
	ethFac, _ := decimal.NewFromString("0.000000000000000001")
	amount := balanceDecimal.Mul(ethFac)
	settingBalance := decimal.NewFromFloat(config.MaxBalance)
	if amount.GreaterThan(settingBalance) {
		fromHex, toHex, rawTxHex, value, err := constructTx(client, *nonce, balance, from, to)
		if err != nil {
			return errors.New(strings.Join([]string{"constructTx error", err.Error()}, " "))
		}
		if err := exportHexTx(fromHex, toHex, rawTxHex, value, nonce, false); err != nil {
			return errors.New(strings.Join([]string{"sub address:", *from, "hased applied withdraw, but fail to export rawTxHex to ", config.RawTx, err.Error()}, " "))
		}
		return nil
	}
	return errors.New("balance not fit the configure")
}

func addressWithAmount(nodeclient *ethclient.Client, esClient *elastic.Client, address string) (*string, *big.Int, *uint64, error) {
	q := elastic.NewBoolQuery()
	q = q.Must(elastic.NewTermQuery("address", address))
	searchResult, _ := esClient.Search().Index("eth_sub_address").Type("sub_address").Query(q).Do(context.Background())

	if len(searchResult.Hits.Hits) < 1 {
		return nil, nil, nil, errors.New(strings.Join([]string{address, "not found in es"}, " "))
	}

	var newSubAddress = new(esSubAddress)

	hit := searchResult.Hits.Hits[0]
	if err := json.Unmarshal(*hit.Source, newSubAddress); err != nil {
		return nil, nil, nil, err
	}
	balance, err := nodeclient.BalanceAt(context.Background(), common.HexToAddress(newSubAddress.Address), nil)
	if err != nil {
		return nil, nil, nil, errors.New(strings.Join([]string{"Failed to get ethereum balance from address:", newSubAddress.Address, err.Error()}, " "))
	}

	pendingNonceAt, _ := nodeclient.PendingNonceAt(context.Background(), common.HexToAddress(newSubAddress.Address))

	return &(newSubAddress.Address), balance, &pendingNonceAt, nil
}

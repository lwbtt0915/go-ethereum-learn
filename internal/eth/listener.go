package eth

import (
	"context"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"gorm.io/gorm"
)

func Start(
	client *ethclient.Client,
	db *gorm.DB,
	start *big.Int,
) {
	query := ethereum.FilterQuery{
		FromBlock: start,
	}

	// 回溯
	logs, _ := client.FilterLogs(context.Background(), query)
	for _, vLog := range logs {
		DecodeAndSave(db, vLog)
	}

	// 实时
	ch := make(chan types.Log)
	sub, _ := client.SubscribeFilterLogs(context.Background(), query, ch)

	for {
		select {
		case err := <-sub.Err():
			log.Println("sub err:", err)
		case vLog := <-ch:
			DecodeAndSave(db, vLog)
		}
	}
}

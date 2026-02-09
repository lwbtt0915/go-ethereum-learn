package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	config "go-ethereum-learn/config"
	"go-ethereum-learn/internal/api"
	"go-ethereum-learn/internal/db"
	"go-ethereum-learn/internal/eth"
	"log"
	"math/big"
)

func main() {
	eth.InitABI()

	//client, err := ethclient.Dial(config.ETH_WS)
	client, err := ethclient.Dial(config.ETH_WS)
	if err != nil {
		log.Fatal(err)
	}

	contractAddress := common.HexToAddress("0x7f52576Bf43A55F533972A238daa2E794cB9b87A")
	transferSig := []byte("Transfer(address,address,uint256)")
	transferTopic := crypto.Keccak256Hash(transferSig)

	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
		Topics:    [][]common.Hash{{transferTopic}},
	}

	logs := make(chan types.Log)

	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)

		case vLog := <-logs:
			from := common.HexToAddress(vLog.Topics[1].Hex())
			to := common.HexToAddress(vLog.Topics[2].Hex())
			value := new(big.Int).SetBytes(vLog.Data)

			fmt.Printf("Transfer %s -> %s : %s\n",
				from.Hex(),
				to.Hex(),
				value.String(),
			)
		}
	}

	database := db.Init(config.MYSQL_DSN)

	//go eth.Start(
	//	client,
	//	database,
	//	big.NewInt(config.START_BLOCK),
	//)

	r := api.Router(database)
	r.Run(":8081")
}

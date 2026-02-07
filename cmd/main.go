package main

import (
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/ethclient"
	config "go-ethereum-learn/config"
	"go-ethereum-learn/internal/api"
	"go-ethereum-learn/internal/db"
	"go-ethereum-learn/internal/eth"
)

func main() {
	eth.InitABI()

	client, err := ethclient.Dial(config.ETH_WS)
	if err != nil {
		log.Fatal(err)
	}

	database := db.Init(config.MYSQL_DSN)

	go eth.Start(
		client,
		database,
		big.NewInt(config.START_BLOCK),
	)

	r := api.Router(database)
	r.Run(":8081")
}

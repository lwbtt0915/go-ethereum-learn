package eth

import (
	"math/big"

	dbm "go-ethereum-learn/internal/db"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"gorm.io/gorm"
)

func DecodeAndSave(db *gorm.DB, l types.Log) {
	switch l.Topics[0] {

	case ERC20ABI.Events["Transfer"].ID:
		var e struct{ Value *big.Int }
		ERC20ABI.UnpackIntoInterface(&e, "Transfer", l.Data)

		db.Create(&dbm.ERC20Event{
			Contract:    l.Address.Hex(),
			EventName:   "Transfer",
			TxHash:      l.TxHash.Hex(),
			LogIndex:    l.Index,
			BlockNumber: l.BlockNumber,
			From:        common.HexToAddress(l.Topics[1].Hex()).Hex(),
			To:          common.HexToAddress(l.Topics[2].Hex()).Hex(),
			Value:       e.Value.String(),
		})

	case ERC721ABI.Events["Transfer"].ID:
		tokenId := new(big.Int).SetBytes(l.Topics[3].Bytes())
		db.Create(&dbm.ERC721Event{
			Contract:    l.Address.Hex(),
			EventName:   "Transfer",
			TxHash:      l.TxHash.Hex(),
			LogIndex:    l.Index,
			BlockNumber: l.BlockNumber,
			From:        common.HexToAddress(l.Topics[1].Hex()).Hex(),
			To:          common.HexToAddress(l.Topics[2].Hex()).Hex(),
			TokenID:     tokenId.String(),
		})
	}
}

package api

import (
	dbm "go-ethereum-learn/internal/db"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Router(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	r.GET("/erc20", func(c *gin.Context) {
		var list []dbm.ERC20Event
		db.Order("block_number desc").Limit(50).Find(&list)
		c.JSON(200, list)
	})

	r.GET("/erc721", func(c *gin.Context) {
		var list []dbm.ERC721Event
		db.Order("block_number desc").Limit(50).Find(&list)
		c.JSON(200, list)
	})

	return r
}

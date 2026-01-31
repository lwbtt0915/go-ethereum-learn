package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 全局变量
var (
	ethClient *ethclient.Client // 以太坊客户端
	db        *gorm.DB          // MySQL数据库连接
	// 合约ABI（示例：USDC的Transfer事件ABI，需替换为你的合约ABI）
	contractABI = `[{"anonymous":false,"inputs":[{"indexed":true,"name":"from","type":"address"},{"indexed":true,"name":"to","type":"address"},{"indexed":false,"name":"value","type":"uint256"}],"name":"Transfer","type":"event"}]`
)

// EventLog 数据库模型：对应contract_events表
type EventLog struct {
	ID              uint            `gorm:"primarykey" json:"id"`
	ContractAddress string          `gorm:"size:42;not null" json:"contract_address"`
	EventName       string          `gorm:"size:100;not null" json:"event_name"`
	TxHash          string          `gorm:"size:66;not null" json:"tx_hash"`
	BlockNumber     int64           `gorm:"not null" json:"block_number"`
	BlockTime       int64           `gorm:"not null" json:"block_time"`
	FromAddress     string          `gorm:"size:42" json:"from_address"`
	EventData       json.RawMessage `gorm:"type:json;not null" json:"event_data"`
	CreatedAt       string          `gorm:"autoCreateTime" json:"created_at"`
}

// 初始化以太坊客户端
func initEthClient() error {
	nodeURL := os.Getenv("ETH_NODE_URL")
	client, err := ethclient.Dial(nodeURL)
	if err != nil {
		return fmt.Errorf("连接以太坊节点失败: %v", err)
	}
	ethClient = client
	return nil
}

// 初始化MySQL数据库
func initMySQL() error {
	dsn := os.Getenv("MYSQL_DSN")
	conn, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("连接MySQL失败: %v", err)
	}
	// 自动迁移表结构（确保表存在）
	conn.AutoMigrate(&EventLog{})
	db = conn
	return nil
}

// 解析合约事件日志
func parseEventLog(log types.Log, abiObj abi.ABI) (map[string]interface{}, string, error) {
	// 遍历ABI中的事件，匹配日志的Topic[0]（事件签名哈希）
	for _, event := range abiObj.Events {
		eventSigHash := abiObj.EventID(event).Hex()
		if log.Topics[0].Hex() == eventSigHash {
			// 解析事件参数
			eventData := make(map[string]interface{})
			if err := abiObj.UnpackIntoMap(eventData, event.Name, log.Data); err != nil {
				return nil, "", fmt.Errorf("解析事件数据失败: %v", err)
			}
			// 解析索引化参数（Topic[1:]）
			for i, input := range event.Inputs {
				if input.Indexed && i < len(log.Topics)-1 {
					val, err := abi.parseParameter(input.Type, log.Topics[i+1].Bytes())
					if err != nil {
						return nil, "", fmt.Errorf("解析索引参数失败: %v", err)
					}
					eventData[input.Name] = val
				}
			}
			return eventData, event.Name, nil
		}
	}
	return nil, "", fmt.Errorf("未匹配到事件")
}

// 监听合约事件并写入数据库
func listenContractEvents() {
	// 1. 解析合约ABI
	abiObj, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		log.Fatalf("解析ABI失败: %v", err)
	}

	// 2. 配置监听参数
	contractAddr := common.HexToAddress(os.Getenv("CONTRACT_ADDRESS"))
	startBlock, _ := strconv.ParseInt(os.Getenv("START_BLOCK"), 10, 64)
	if startBlock == 0 {
		// 获取最新区块高度，从最新区块开始监听
		latestBlock, err := ethClient.BlockNumber(context.Background())
		if err != nil {
			log.Fatalf("获取最新区块失败: %v", err)
		}
		startBlock = int64(latestBlock)
		log.Printf("从最新区块 %d 开始监听", startBlock)
	}

	// 3. 构建日志查询过滤器
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddr}, // 仅监听指定合约
		FromBlock: big.NewInt(startBlock),
		ToBlock:   nil, // nil=监听最新区块（实时）
	}

	// 4. 启动事件监听
	logs := make(chan types.Log)
	sub, err := ethClient.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.Fatalf("订阅事件失败: %v", err)
	}

	log.Println("开始监听合约事件...")
	for {
		select {
		case err := <-sub.Err():
			log.Printf("订阅出错: %v，重新订阅...", err)
			// 出错后重新订阅
			sub, _ = ethClient.SubscribeFilterLogs(context.Background(), query, logs)
		case vLog := <-logs:
			// 解析事件
			eventData, eventName, err := parseEventLog(vLog, abiObj)
			if err != nil {
				log.Printf("解析日志失败: %v", err)
				continue
			}

			// 获取区块时间
			block, err := ethClient.BlockByHash(context.Background(), vLog.BlockHash)
			if err != nil {
				log.Printf("获取区块信息失败: %v", err)
				continue
			}

			// 构建事件数据
			eventJSON, _ := json.Marshal(eventData)
			eventLog := EventLog{
				ContractAddress: contractAddr.Hex(),
				EventName:       eventName,
				TxHash:          vLog.TxHash.Hex(),
				BlockNumber:     int64(vLog.BlockNumber),
				BlockTime:       block.ReceivedAt.UnixNano(),
				FromAddress:     "", // 可从交易中解析发起者地址，示例省略
				EventData:       eventJSON,
			}

			// 写入数据库
			if err := db.Create(&eventLog).Error; err != nil {
				log.Printf("写入数据库失败: %v", err)
			} else {
				log.Printf("成功写入事件: %s, 交易哈希: %s", eventName, vLog.TxHash.Hex())
			}
		}
	}
}

// 初始化路由（查询接口）
func initRouter() *gin.Engine {
	r := gin.Default()

	// 事件查询接口
	api := r.Group("/api/v1/eth/events")
	{
		// 按合约地址+事件名称查询
		api.GET("", func(c *gin.Context) {
			contractAddr := c.Query("contract_address")
			eventName := c.Query("event_name")
			blockNumber := c.Query("block_number")
			page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
			limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
			offset := (page - 1) * limit

			// 构建查询条件
			query := db.Model(&EventLog{})
			if contractAddr != "" {
				query = query.Where("contract_address = ?", contractAddr)
			}
			if eventName != "" {
				query = query.Where("event_name = ?", eventName)
			}
			if blockNumber != "" {
				query = query.Where("block_number = ?", blockNumber)
			}

			// 分页查询
			var events []EventLog
			var total int64
			query.Count(&total)
			query.Order("block_number DESC").Offset(offset).Limit(limit).Find(&events)

			c.JSON(http.StatusOK, gin.H{
				"code": 0,
				"msg":  "success",
				"data": gin.H{
					"list":  events,
					"total": total,
					"page":  page,
					"limit": limit,
				},
			})
		})

		// 按交易哈希查询单个事件
		api.GET("/tx/:tx_hash", func(c *gin.Context) {
			txHash := c.Param("tx_hash")
			var event EventLog
			if err := db.Where("tx_hash = ?", txHash).First(&event).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{
					"code": 1,
					"msg":  "事件不存在",
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"code": 0,
				"msg":  "success",
				"data": event,
			})
		})
	}

	return r
}

func main() {
	// 加载配置文件
	if err := godotenv.Load(); err != nil {
		log.Fatalf("加载.env文件失败: %v", err)
	}

	// 初始化以太坊客户端和MySQL
	if err := initEthClient(); err != nil {
		log.Fatalf("初始化以太坊客户端失败: %v", err)
	}
	if err := initMySQL(); err != nil {
		log.Fatalf("初始化MySQL失败: %v", err)
	}

	// 启动事件监听（异步goroutine）
	go listenContractEvents()

	// 启动HTTP查询服务
	router := initRouter()
	port := os.Getenv("SERVER_PORT")
	log.Printf("查询服务启动成功，端口: %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("启动HTTP服务失败: %v", err)
	}
}

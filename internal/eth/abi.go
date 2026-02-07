package eth

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"strings"
)

var ERC20ABI abi.ABI
var ERC721ABI abi.ABI

func InitABI() {
	ERC20ABI, _ = abi.JSON(strings.NewReader(erc20ABI))
	ERC721ABI, _ = abi.JSON(strings.NewReader(erc721ABI))
}

const erc20ABI = `[
 {"anonymous":false,"inputs":[
  {"indexed":true,"name":"from","type":"address"},
  {"indexed":true,"name":"to","type":"address"},
  {"indexed":false,"name":"value","type":"uint256"}],
  "name":"Transfer","type":"event"},
 {"anonymous":false,"inputs":[
  {"indexed":true,"name":"owner","type":"address"},
  {"indexed":true,"name":"spender","type":"address"},
  {"indexed":false,"name":"value","type":"uint256"}],
  "name":"Approval","type":"event"}
]`

const erc721ABI = `[
 {"anonymous":false,"inputs":[
  {"indexed":true,"name":"from","type":"address"},
  {"indexed":true,"name":"to","type":"address"},
  {"indexed":true,"name":"tokenId","type":"uint256"}],
  "name":"Transfer","type":"event"},
 {"anonymous":false,"inputs":[
  {"indexed":true,"name":"owner","type":"address"},
  {"indexed":true,"name":"approved","type":"address"},
  {"indexed":true,"name":"tokenId","type":"uint256"}],
  "name":"Approval","type":"event"},
 {"anonymous":false,"inputs":[
  {"indexed":true,"name":"owner","type":"address"},
  {"indexed":true,"name":"operator","type":"address"},
  {"indexed":false,"name":"approved","type":"bool"}],
  "name":"ApprovalForAll","type":"event"}
]`

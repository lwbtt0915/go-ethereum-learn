package db

type ERC20Event struct {
	ID          uint `gorm:"primaryKey"`
	Contract    string
	EventName   string
	TxHash      string `gorm:"uniqueIndex:uk20"`
	LogIndex    uint   `gorm:"uniqueIndex:uk20"`
	BlockNumber uint64
	From        string
	To          string
	Value       string
}

type ERC721Event struct {
	ID          uint `gorm:"primaryKey"`
	Contract    string
	EventName   string
	TxHash      string `gorm:"uniqueIndex:uk721"`
	LogIndex    uint   `gorm:"uniqueIndex:uk721"`
	BlockNumber uint64
	From        string
	To          string
	TokenID     string
	Operator    string
	Approved    *bool
}

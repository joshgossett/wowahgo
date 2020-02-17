package wowdb

import (
	"time"

	"github.com/jinzhu/gorm"
)

//Auction contains infomation about an auction listing
type Auction struct {
	//ID         int64 `gorm:"primary_key"`
	gorm.Model
	DumpFileID uint
	AuctionID  int64  `json:"auc"`
	ItemID     int64  `json:"item"`
	Owner      string `json:"owner"`
	OwnerRealm string `json:"ownerRealm"`
	Bid        int64  `json:"bid"`
	Buyout     int64  `json:"buyout"`
	Quantity   int    `json:"quantity"`
	TimeLeft   string `json:"timeLeft"`
	Rand       int64  `json:"rand"`
	Seed       int64  `json:"seed"`
	Context    int64  `json:"context"`
}

//Auctions has a list of Auctions
type Auctions struct {
	//ID       int64 `gorm:"primary_key"`
	gorm.Model
	Auctions []Auction
}

//DumpFiles is a list of files in the AH dump
type DumpFiles struct {
	Files []DumpFile
}

//DumpFile contains an URL to the AH dump and a time it was dumped
type DumpFile struct {
	//ID           int64  `gorm:"primary_key"`
	gorm.Model
	URL          string //`json:"url"`
	LastModified int64  //`json:"lastModified"`
}

//GetTime returns the Time of this dump
func (f *DumpFile) GetTime() time.Time {
	secs := f.LastModified / 1000
	return time.Unix(secs, 0)
}

package wowdb

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/jinzhu/gorm"
)

type Config struct {
	ConnectionString string `json:"connectionString"`
	APIKey           string `json:"apiKey"`
	APISecret        string `json:"apiSecret"`
	Driver           string `json:"sqldriver"`
	MaxDBConnections int    `json:"maxDBConnections"`
	TimeoutMins      int    `json:"timeoutMins"`
}

func LoadConfig() (*Config, error) {
	log.Println("opening config file...")
	f, err := os.Open("config.json")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var conf Config
	err = json.Unmarshal(b, &conf)
	if err != nil {
		return nil, err
	}
	log.Println("Config file loaded")
	return &conf, nil
}

func OpenDB(driver string, connectionString string, maxConnections int) (*gorm.DB, error) {
	db, err := gorm.Open(driver, connectionString)
	if err != nil {
		log.Println("Error opening connection to database: ", err)
		return nil, err
	}

	db.DB().SetMaxIdleConns(maxConnections)
	db.DB().SetMaxOpenConns(maxConnections)
	return db, err
}

//GetDumpInfo returns the DUmpFiles for thegiven realm
func GetDumpInfo(realm string, apiKey string) (*DumpFiles, error) {
	res, err := http.Get("https://us.api.battle.net/wow/auction/data/" + realm + "?locale=en_US&apikey=" + apiKey)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, errors.New("Status code: " + strconv.Itoa(res.StatusCode))
	}
	val, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var f DumpFiles
	json.Unmarshal(val, &f)

	log.Println("got", len(f.Files), "auction dump files in list")
	for _, v := range f.Files {
		log.Println("URL: ", v.URL)
		log.Println("DUMP TIME: ", v.GetTime())
	}
	return &f, nil
}

//GetAHDump returns the list of auctions from the given dump file url
func GetAHDump(url string) (*Auctions, error) {
	res, err := http.Get(url)
	if err != nil {
		log.Println("Error retreiving AH file from server: ", err)
		return nil, err
	}
	if res.StatusCode != 200 {
		log.Println("Wrong status code: ", res.StatusCode)
		return nil, errors.New("Wrong status code: " + strconv.Itoa(res.StatusCode))
	}
	val, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("Error reading body: ", err)
		return nil, err
	}
	var a Auctions
	log.Println("Parsing auction data...")
	json.Unmarshal(val, &a)
	log.Println("Finished parsing json data")

	return &a, nil
}

//BuildDBDebug builds the database
func InstallDB(db *gorm.DB) {
	log.Println("Creating database tables...")
	err := db.CreateTable(&DumpFile{}).Error
	if err != nil {
		panic(err)
	}

	err = db.CreateTable(&Auction{}).Error
	if err != nil {
		panic(err)
	}
	log.Println("Database tables created")
}

func UninstallDB(db *gorm.DB) {
	log.Println("Dropping databse tables...")
	db.DropTable(&Auction{})
	db.DropTable(&DumpFile{})
	log.Println("Finished dropping database tables")
}

//InsertAuction inserts a new auction into the table
func InsertAuction(auc *Auction, db *gorm.DB) error {
	err := db.Create(auc).Error
	return err
}

//Returns the primary key of the dumpfile it inserted
func InsertDump(dmpFile DumpFile, db *gorm.DB) (uint, error) {
	err := db.Create(&dmpFile).Error
	if err != nil {
		log.Println("error inserting dump file: ", dmpFile.URL)
		return 0, err
	}
	return dmpFile.ID, nil
}

func DoesFileExist(dmp DumpFile, db *gorm.DB) (bool, error) {
	count := 0
	err := db.Model(&DumpFile{}).Where("last_modified = ?", dmp.LastModified).Count(&count).Error
	if err != nil {
		log.Println(err)
		return true, err
	}
	if count > 0 {
		return true, nil
	} else {
		return false, nil
	}
}

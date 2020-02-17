package main

import (
	"flag"
	"log"
	"os"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/joshgossett/wowah/wowdb"
	_ "github.com/mattn/go-sqlite3"
)

var wg sync.WaitGroup

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Println("recovered from error: ", err)
		}
	}()

	var install bool
	var uninstall bool
	var debug bool
	flag.BoolVar(&install, "install", false, "Installs the database schema")
	flag.BoolVar(&uninstall, "uninstall", false, "Uninstalls the database schema")
	flag.BoolVar(&debug, "debug", false, "Sets the output of the program to the console instead of to a file")
	flag.Parse()

	if !debug {
		f, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			log.Println("Failed to open output file")
		}
		log.SetOutput(f)
		defer f.Close()
	}

	conf, err := wowdb.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	//install this program and exit
	if install {
		db, err := wowdb.OpenDB(conf.Driver, conf.ConnectionString, conf.MaxDBConnections)
		if err != nil {
			log.Fatal(err)
		}
		wowdb.InstallDB(db)
		db.Close()
		return
	}

	//uninstall this program and exit
	if uninstall {
		db, err := wowdb.OpenDB(conf.Driver, conf.ConnectionString, conf.MaxDBConnections)
		if err != nil {
			log.Fatal(err)
		}
		wowdb.UninstallDB(db)
		db.Close()
		return
	}

	RunAHLoop("korgath", conf)

}

func RunAHLoop(realm string, conf *wowdb.Config) {

	timer := time.NewTicker(time.Minute * time.Duration(conf.TimeoutMins))
	for {
		func() {
			log.Println("Opening DB connection...")
			db, err := wowdb.OpenDB(conf.Driver, conf.ConnectionString, conf.MaxDBConnections)
			if err != nil {
				log.Fatal(err)
			}
			log.Println("DB connection open")
			//execute at end of every scope
			defer func() {
				db.Close()
				log.Println("DB connection closed")
			}()
			file, err := wowdb.GetDumpInfo(realm, conf.APIKey)
			if err != nil {
				log.Fatal(err)
			}

			for _, v := range file.Files {
				start := time.Now()
				var stop time.Time
				log.Println("Inserting ah dump...")

				if t, err := wowdb.DoesFileExist(v, db); t {
					log.Println("File already exists in database. Continuing...")
					continue
				} else if err != nil {
					log.Fatal(err)
				}

				dID, err := wowdb.InsertDump(v, db)
				if err != nil {
					log.Fatal(err)
				}

				log.Println("Inserted dump with ID:", dID)
				log.Println("Downloading ah data...")

				auc, err := wowdb.GetAHDump(v.URL)
				if err != nil {
					log.Fatal(err)
				}
				log.Println("Finished downloading ah data")

				log.Println("inserting auctions into DB...")
				for _, a := range auc.Auctions {
					wg.Add(1)
					a.DumpFileID = dID
					go func(a wowdb.Auction, db *gorm.DB) {
						defer wg.Done()
						err := wowdb.InsertAuction(&a, db)
						if err != nil {
							log.Println("Failed to insert auction: ", err)
						}
					}(a, db)
				}
				wg.Wait()
				stop = time.Now()
				total := stop.Sub(start)
				log.Println("finished insterting into DB after:", total)
				log.Println("inserted ", len(auc.Auctions), " items")
			}
			log.Println("Finished checking all files for this dataset.")
		}()
		<-timer.C
	}

}

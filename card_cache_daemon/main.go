package main;

import (
	"fmt"
	"os"
	"time"
	"log"
	"io"
  "github.com/joho/godotenv"
  "gorm.io/driver/postgres"
  "gorm.io/gorm"
  "net/http"
  "crypto/tls"
	//"encoding/json"
)

type Config struct {
	dbName string
	dbPassword string
	dbUsername string
	dbUrl string
	dbPort string
	dataSource string
}

func connect(config Config) (*gorm.DB, error) {
  db, err := gorm.Open(postgres.New(postgres.Config{
    DSN: "host=" + config.dbUrl + " user=" + config.dbUsername + " password=" + config.dbPassword + " dbname=" + config.dbName + " port=" + config.dbPort, 
    PreferSimpleProtocol: true,}), &gorm.Config{});

  return db, err
}

func update_internal(db *gorm.DB, config Config) bool {
	fmt.Println(db); // I use this go, so go put it up your arse

	log.Println("Updating the card cache...");
  client := &http.Client{
    Transport: &http.Transport{
      TLSClientConfig: &tls.Config{
      },
    },
  }

  resp, err := client.Get(config.dataSource)

	if (err != nil) {
    log.Println("Could not fetch cards from source");
    return false;
  }

	log.Println("Download started...");
	body, err := io.ReadAll(resp.Body);
  if (err != nil) {
  	log.Println("An error occured reading the body");
  	return false;
  }

	log.Printf("Download finished, body length: %d\n", len(body));
  log.Println("Parsing cards...");
  return true;
}

const MAX = 10;

func update(db *gorm.DB, config Config) {
	i := int (0);
  for (i < MAX) {
  	if (i > 0) {
  		log.Printf("Trying to fetch cards again %d/%d\n", i + 1, MAX);
  	} else {
  		log.Println("Trying to fetch the card cache...");
  	}

  	if (update_internal(db, config)) {
      log.Println("Update successful.");
  		break;
  	} else {
      log.Println("Failed to fetch cards");
		}

		i++;
  }
}

const WAIT_TIME = time.Millisecond * 1000 * 60 * 60 * 12;
const REPO = "https://github.com/MonarchDevelopment/MonarchDB";
const VERSION = "V1.0.0";

func main() {
  fmt.Println("Loading MonarchDB Card Cache Daemon");
  fmt.Printf(" -> Version %s | Repo %s\n", VERSION, REPO);

	// Get environment
	godotenv.Load()
	dbname := os.Getenv("DB_NAME")
	dbpassword := os.Getenv("DB_PASSWORD")
	dbusername := os.Getenv("DB_USERNAME")
	dburl := os.Getenv("DB_URL")
	dbport := os.Getenv("DB_PORT")
	datasource := os.Getenv("DATA_SOURCE")

	// Test for empty vars
	if (dbname == "") {
    panic("DB_NAME is not defined");
	}
	if (dbpassword == "") {
    panic("DB_PASSWORD is not defined");
  }
  if (dbusername == "") {
    panic("DB_USERNAME is not defined");
  }
  if (dburl == "") {
    panic("DB_URL is not defined");
  }
  if (datasource == "") {
    panic("DATA_SOURCE is not defined, the default is https://mtgjson.com/api/v5/AtomicCards.json");
  }

	// Put in config
	config := Config {
					dbName: dbname,
					dbPassword: dbpassword,
					dbUsername: dbusername,
					dbUrl: dburl,
					dbPort: dbport,
					dataSource: datasource};

	fmt.Printf("Configuration: %s\n", config);
	fmt.Println("Testing database connection...");

	db, err := connect(config);
	if (err != nil) {
		panic("Cannot connect to the database");
	}

	log.SetFlags(2 | 3);
	log.Println("Connection successful, starting daemon.");

	lastupdate := time.Now();
	for true {
	  lastupdate = time.Now();
		update(db, config);
		
		log.Println("Waiting for the next update.");
		diff := time.Now().Sub(lastupdate).Nanoseconds();
		for (diff < WAIT_TIME.Nanoseconds()) {
			time.Sleep(time.Millisecond * 100);
  		diff = time.Now().Sub(lastupdate).Nanoseconds();
		}
	}
}


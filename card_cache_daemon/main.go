package main

import (
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Config struct {
	dbName     string
	dbPassword string
	dbUsername string
	dbUrl      string
	dbPort     string
}

type Card struct {
	OracleId       string `json:"uuid"`
	CardName       string `json:"name"`
	OracleText     string `json:"text,omitempty"`
	scryfallUri    string
	Colour         []string  `json:"colors"`
	ColourIdentity []string  `json:"colorIdentity"`
	Type           string  `json:"types"`
	Cmc            float64 `json:"convertedManaCost"`
	ManaCost       string  `json:"manaCost,omitempty"`
}

/*
create table cards (
oracle_id uuid primary key,
scryfall_uri varchar(512) not null,
card_name varchar(255) not null,
color varchar(255) not null,
color_identity varchar(255) not null,
type varchar(255) not null,
cmc double precision not null,
mana_cost varchar(255) not null,
oracle_text varchar(1024) not null
);
*/

const JSON_URI="https://c2.scryfall.com/file/scryfall-bulk/oracle-cards/oracle-cards-20220403090406.json"

func getScryfallUri(card Card) string {
	nameenc := url.QueryEscape(card.CardName)
	return "https://scryfall.com/search?q=name%3D%2F%5E" + nameenc + "%24%2F&unique=cards&as=grid&order=name"
}

func connect(config Config) (*sql.DB, error) {
	db, err := sql.Open("postgres",
		fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
			config.dbUsername,
			config.dbPassword,
			config.dbUrl,
			config.dbPort,
			config.dbName))

	return db, err
}

func insert_cards(_ *sql.DB, _ Config, data []Card) error {
	// Get local cache
	log.Println(data)

	return nil
}

func update_internal(db *sql.DB, config Config) bool {
	log.Println("Updating the card cache...")
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{},
		},
	}

	resp, err := client.Get(JSON_URI)

	if err != nil {
		log.Println(err)
		log.Println("Could not fetch cards from source")
		return false
	}

	log.Println("Download started...")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		log.Println("An error occured reading the body")
		return false
	}
	defer resp.Body.Close()

	log.Printf("Download finished, body length: %d\n", len(body))
	log.Println("Parsing cards...")

	var data []Card
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Println(err)
		log.Println("An error occured parsing the cards")
		return false
	}

	err = insert_cards(db, config, data)
	if err != nil {
		return false
	}

	return true
}

const MAX = 10

func update(db *sql.DB, config Config) {
	i := int(0)
	for i < MAX {
		if i > 0 {
			log.Printf("Trying to fetch cards again %d/%d\n", i+1, MAX)
		} else {
			log.Println("Trying to fetch the card cache...")
		}

		if update_internal(db, config) {
			log.Println("Update successful.")
			break
		} else {
			log.Println("Failed to fetch cards")
		}

		i++
	}
}

const WAIT_TIME = time.Millisecond * 1000 * 60 * 60 * 12
const REPO = "https://github.com/MonarchDevelopment/MonarchDB"
const VERSION = "V1.0.0"

func main() {
	fmt.Println("Loading MonarchDB Card Cache Daemon")
	fmt.Printf(" -> Version %s | Repo %s\n", VERSION, REPO)

	// Get environment
	godotenv.Load()
	dbname := os.Getenv("DB_NAME")
	dbpassword := os.Getenv("DB_PASSWORD")
	dbusername := os.Getenv("DB_USERNAME")
	dburl := os.Getenv("DB_URL")
	dbport := os.Getenv("DB_PORT")

	// Test for empty vars
	if dbname == "" {
		panic("DB_NAME is not defined")
	}
	if dbpassword == "" {
		panic("DB_PASSWORD is not defined")
	}
	if dbusername == "" {
		panic("DB_USERNAME is not defined")
	}
	if dburl == "" {
		panic("DB_URL is not defined")
	}

	// Put in config
	config := Config{
		dbName:     dbname,
		dbPassword: dbpassword,
		dbUsername: dbusername,
		dbUrl:      dburl,
		dbPort:     dbport}

	fmt.Printf("Configuration: %s\n", config)
	fmt.Println("Testing database connection...")

	db, err := connect(config)
	if err != nil {
		panic("Cannot connect to the database")
	}

	log.SetFlags(2 | 3)
	log.Println("Connection successful, starting daemon.")

	lastupdate := time.Now()
	for true {
		lastupdate = time.Now()
		update(db, config)

		log.Println("Waiting for the next update.")
		diff := time.Now().Sub(lastupdate).Nanoseconds()
		for diff < WAIT_TIME.Nanoseconds() {
			time.Sleep(time.Millisecond * 100)
			diff = time.Now().Sub(lastupdate).Nanoseconds()
		}
	}
}

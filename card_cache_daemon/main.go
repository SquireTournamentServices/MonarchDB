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

const MDFC = "modal_dfc"
const FLIP = "flip"
const TRANSFORM = "transform"
const ACCEPTED_FACE = ""

// this or empty string are valid,
// all cards that have invalid faces are put into a queue and then they are added to the
// card oracle at the end. to make sure they are unique a composite key in a map is used
// of `name + face`

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
	Latouts        string `json:"layout,omitempty"`
	scryfallUri    string
	Colour         []string `json:"colors"`
	ColourIdentity []string `json:"colorIdentity"`
	Type           []string `json:"types"`
	Cmc            float64  `json:"convertedManaCost"`
	ManaCost       string   `json:"manaCost,omitempty"`
	Face           string   `json:"face,omitempty"`
}

type Set struct {
	Cards []Card `json:"cards"`
}

type AllPrintings struct {
	Sets map[string]Set `json:"data"`
}

/*
create table cards (
cardid uuid primary key,
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
const SQL_GET_CARDS = "select (cardid) from cards;"
const JSON_URI = "https://mtgjson.com/api/v5/AllPrintings.json"

func getScryfallUri(card Card) string {
	nameenc := url.QueryEscape(card.CardName)
	return "https://scryfall.com/search?q=name%3D%2F%5E" + nameenc + "%24%2F&unique=cards&as=grid&order=name"
}

func connect(config Config) (*sql.DB, error) {
	db, err := sql.Open("postgres",
		fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
			config.dbUsername,
			config.dbPassword,
			config.dbUrl,
			config.dbPort,
			config.dbName))

	return db, err
}

const INSERT_CARD_SQL = "insert (cardid, scryfall_uri, card_name, color, color_identity, type, cmc, mana_cost, oracle_text into cards values ($1, $2, $3, $4, $5, $6, $7, $8 ,$9);"
const UPDATE_CARD_SQL = "update cards set scryfall_uri=$1, card_name=$2, color=$3, color_identity=$4, type=$5, cmc=$6, mana_cost=$7, oracle_text=$8;"

func insert_cards(db *sql.DB, _ Config, data []Card) error {
	log.Println("Syncing card database")
	// Get all cards in local cache and add to a map
	var oracleIdMap map[string]bool = make(map[string]bool)

	rows, err := db.Query(SQL_GET_CARDS)
	if err != nil {
		log.Println(err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var oracleId string
		err = rows.Scan(&oracleId)
		if err != nil {
			log.Println(err)
			return err
		}
		oracleIdMap[oracleId] = true
	}

	err = rows.Err()
	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("Performing updates")
	// Iterate over all cards in the database and insert/update if needed.
	inserts := 0
	updates := 0

	for i := 0; i < len(data); i++ {
		_, inMap := oracleIdMap[data[i].OracleId]
		if !inMap {
			log.Printf("Inserting %s\n", data[i].CardName)
			inserts++

			// Insert
		} /*else {
		  // Check for update
		  card := oracleIdMap[data[i].OracleId]
		}*/
	}

	log.Printf("Inserted %d cards, Updated %d cards\n", inserts, updates)

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

	var data AllPrintings
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Println(err)
		log.Println("An error occured parsing the cards")
		return false
	}

	log.Printf("Found %d sets\n", len(data.Sets))
	log.Println("Merging identical cards...\n")

	var cards_map map[string]Card = make(map[string]Card)
	i := 0
	for key, val := range data.Sets {
		log.Printf("Processing %s - %d/%d\n", key, i, len(data.Sets))
		for _, card := range val.Cards {
			_, found := cards_map[card.CardName]
			if !found {
				if card.Face == ACCEPTED_FACE {
					cards_map[card.CardName] = card
				}
			}
		}

		i += 1
	}

	var cards []Card = make([]Card, 0)
	for _, val := range cards_map {
		val.scryfallUri = getScryfallUri(val)
		cards = append(cards, val)
	}

	log.Println("Inserting / updating cards")

	err = insert_cards(db, config, cards)
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
	fmt.Println(" -> Licenced under GPL 3 for use freely by all :)")

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

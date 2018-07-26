package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"syscall"
	"text/tabwriter"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/ssh/terminal"

	dataStruct "./DataGets/DataStruct"
	elwatch "./DataGets/ElwatchGet"
	ttn "./DataGets/TTNGet"
	telenor "./DataGets/TelenorGet"
	ubiik "./DataGets/UbiikGet"
	secrets "./secrets"
)

type fn func(string, secrets.Info) (dataStruct.Data, error)

func pullData(sensorID, name string, secret secrets.Info, getFunction fn, channel chan (dataStruct.Data)) {

	for {
		time.Sleep(1 * time.Second)
		data, err := getFunction(sensorID, secret)
		if err != nil {
			fmt.Printf("\n%s Error: %s\n", name, err.Error())
		} else if lastValue[data.ID].String() != data.Time.String() {
			channel <- data
			lastValue[data.ID] = data.Time
		}
	}
}

func pullUbiikData(sensorID string) {
	pullData(sensorID, "Ubiik", secrets.Ubiik, ubiik.GetUbiikData, channelUbiik)
}

func pullTTNData(sensorID string) {
	pullData(sensorID, "TTN", secrets.TTN, ttn.GetTTNData, channelTTN)
}

func pullTelenorData(sensorID string) {
	pullData(sensorID, "Telenor", secrets.Telenor, telenor.GetTelenorData, channelTelenor)
}

func pullElwatchData(sensorID string) {
	pullData(sensorID, "Elwatch", secrets.Elwatch, elwatch.GetElwatchData, channelElwatch)
}

func pushData() {
	db, err := sql.Open("mysql", "elias:"+password+"@tcp(lowpowersensor.tk:3306)/elwatch")
	if err != nil {
		log.Println("write:", err)
	}

	defer db.Close()

	for {
		time.Sleep(1 * time.Second)
		data := <-channelPush
		addMYSQLData(data, db)
	}
}

func addMYSQLData(collection []dataStruct.Data, db *sql.DB) {

	for _, data := range collection {

		// Try to convert to float
		_, err := strconv.ParseFloat(data.Data, 64)
		if err != nil {
			fmt.Printf("\nCannot convert '%v' to float\n", data.Data)
			continue
		}

		// Add datapoint
		rows, err := db.Query(fmt.Sprintf("INSERT IGNORE INTO sensor_%v (value, date) VALUES (%v, '%v');", data.ID, data.Data, data.Time.Format("2006-01-02 15:04:05")))
		if err != nil {
			// Try to create table
			temp2, err := db.Query(fmt.Sprintf("CREATE TABLE sensor_%v LIKE template;", data.ID))
			fmt.Println("\nCreating table...")
			if err != nil {
				fmt.Println("\nUnexpected error!: " + err.Error())
			}
			temp2.Close()
		}
		rows.Close()
	}
}

func collectData() {
	var dataCollection []dataStruct.Data
	// Initial pull
	for _, id := range sensorIDs.TTN {
		go pullTTNData(id)
	}
	for _, id := range sensorIDs.Telenor {
		go pullTelenorData(id)
	}

	for _, id := range sensorIDs.Elwatch {
		go pullElwatchData(id)
	}

	var TelenorUpdates = 0
	var ElwatchUpdates = 0
	var TTNUpdates = 0
	var start = time.Now()
	var lastPush = time.Now()
	var lastPushPtr = &lastPush
	var timeSincePush = func() time.Duration {
		return time.Now().Sub(*lastPushPtr).Round(time.Second)
	}
	var timeSinceStart = func() time.Duration {
		return time.Now().Sub(start).Round(time.Second)
	}

	fmt.Println("ITP: Items to push \nTSP: Time since push \nTSS: Time since start")

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 10, 10, 0, '\t', tabwriter.Debug|tabwriter.AlignRight)
	fmt.Fprintln(w, "TEL items\tELW items\tTTN items\tITP\tTSP\tTSS")
	var printUpdate = func() {
		fmt.Fprintf(w, "\r%8v\t%8v\t%8v\t%3v\t%3v\t%10v", TelenorUpdates, ElwatchUpdates, TTNUpdates, len(dataCollection), timeSincePush(), timeSinceStart())
		w.Flush()
	}
	for {
		select {
		case data := <-channelTelenor:
			TelenorUpdates++
			printUpdate()
			dataCollection = append(dataCollection, data)

		case data := <-channelTTN:
			TTNUpdates++
			printUpdate()
			dataCollection = append(dataCollection, data)

		case data := <-channelElwatch:
			ElwatchUpdates++
			printUpdate()
			dataCollection = append(dataCollection, data)

		case channelPush <- dataCollection:
			printUpdate()
			if len(dataCollection) != 0 {
				lastPush = time.Now()
			}
			dataCollection = nil
		}
	}
}

var channelTTN = make(chan dataStruct.Data, 10)
var channelTelenor = make(chan dataStruct.Data, 10)
var channelElwatch = make(chan dataStruct.Data, 10)
var channelUbiik = make(chan dataStruct.Data, 10)
var channelPush = make(chan []dataStruct.Data)

var sensorIDs struct {
	Elwatch []string

	Telenor []string

	TTN []string
}

var lastValue = make(map[string]time.Time)

var password string

func main() {
	//pullUbiikData("something")

	sensorIDs.Elwatch = []string{"20006040", "20006039", "20004700", "20005880", "20005883", "20004722", "20004936", "20004874"}
	sensorIDs.Telenor = []string{"357517080049085"}
	sensorIDs.TTN = []string{"temp_reader1"}

	fmt.Print("Enter password for mysql user elias at 46.101.29.167: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err == nil {
	}
	password = string(bytePassword)
	fmt.Println("")
	fmt.Println("Starting data pusher...")
	go pushData()
	fmt.Println("Listening...")
	collectData()

}

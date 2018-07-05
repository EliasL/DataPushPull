package main

import (
	"fmt"
	"time"

	elwatchGet "./DataGets/ElwatchGet"
	secrets "./secrets"
	nbiotget "github.com/NB-IoT_Telenor/NbiotGet"
	loraget "gitlab.com/leskraas/LoRa_TTN/LoRaGet"
)

func pullTTNData(sensorID string, TTNDone chan bool) {
	channelTTN <- loraget.GetTTNData(sensorID, secrets.TTN.APIKey)
	TTNDone <- true
}

func pullTelenorData(sensorID string, TelenorDone chan bool) {
	var url = fmt.Sprintf("https://in.nbiot.engineering/devices/%s/data", sensorID)
	channelTelenor <- nbiotget.GetLastData(url, secrets.Telenor.Username, secrets.Telenor.Password)
	TelenorDone <- true
}

func pullData() {
	TTNDone, TelenorDone, ElwatchDone := make(chan bool), make(chan bool), make(chan bool)
	go pullTTNData("power_compare", TTNDone)
	go pullTelenorData("357517080049085", TelenorDone)
	for {
		select {
		case <-TTNDone:
			go pullTTNData("power_compare", TTNDone)
		case <-TelenorDone:
			go pullTelenorData("357517080049085", TelenorDone)
		case <-ElwatchDone:
			fmt.Println("El-watch: " + <-channelElwatch)

		default:
			fmt.Println("Waiting...")
			time.Sleep(1 * time.Second)
		}
	}
}

func listen() {
	for {
		select {
		case <-channelTelenor:
			fmt.Println("Telenor: " + (<-channelTelenor).Unixtime.String())
		case <-channelTTN:
			fmt.Println("TTN: " + (<-channelTTN).Unixtime.String())
		case <-channelElwatch:
			fmt.Println("El-watch: " + (<-channelElwatch).Unixtime.String())
		}
	}
}

var channelTTN = make(chan DataDecoded)
var channelTelenor = make(chan DataDecoded)
var channelElwatch = make(chan DataDecoded)

func main() {
	d := elwatchGet.GetElwatchData("20006040", secrets.Elwatch.APIKey)
	fmt.Println(d.Unixtime)
	fmt.Println("Starting DataPullers...")
	go pullData()
	fmt.Println("Listening...")
	listen()

}

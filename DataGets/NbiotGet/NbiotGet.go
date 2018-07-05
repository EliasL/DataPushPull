package nbiotget

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

//NanoToSec converts nano to sec. The Timestamp is in nanosec, so devide the timestampe with NanoToSec to get sec
const NanoToSec = 1e9

// Nbiot holds the whole json objekt
type Nbiot struct {
	Data []Data `json:"data"`
}

// NDataDecoded holds the N last DataDecoded
type NDataDecoded struct {
	NData []Data
}

// GetAllRawData returns a Nbiot struct with the whole json object.
func GetAllRawData(url string, username string, passwd string) Nbiot {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(username, passwd)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var d Nbiot
	for {
		if err := decoder.Decode(&d); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
	}

	return d
}

// GetLastData returns the last data in a Data struct
func GetLastData(url string, username string, passwd string) dataStruct.Data {
	d := GetAllRawData(url, username, passwd)
	var data dataStruct.Data
	data.ID = d.Data[len(d.Data)-1].Imei
	temp := d.Data[len(d.Data)-1].Timestamp / NanoToSec
	data.Time = time.Unix(int64(temp), 0)
	data.Data = string(d.Data[len(d.Data)-1].Payload)

	return data
}

// GetLastNData returns the n last data in a NDataDecoded struct
func GetLastNData(url string, username string, passwd string, n int) NDataDecoded {
	d := GetAllRawData(url, username, passwd)
	var ndata NDataDecoded
	for n > 0 {
		var data DataDecoded
		data.ID = d.Data[len(d.Data)-1-n].Imei
		temp := d.Data[len(d.Data)-1-n].Timestamp / NanoToSec
		data.Time = time.Unix(int64(temp), 0)
		data.Data = string(d.Data[len(d.Data)-1-n].Payload)
		ndata.NData = append(ndata.NData, data)
		n--
	}
	return ndata
}

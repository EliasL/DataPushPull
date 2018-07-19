package telenorget

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	secret "../../secrets"
	dataStruct "../DataStruct"
)

//NanoToSec converts nano to sec. The Timestamp is in nanosec, so devide the timestampe with NanoToSec to get sec
const NanoToSec = 1e9

// Data is the temporary response format
type Data struct {
	Imei      string `json:"imei"`
	Timestamp int    `json:"timestamp"`
	Payload   []byte `json:"payload"` // When decoder.Decode gets a []byte, it decodes the value as base64 and returns a ascii array
}

// Nbiot holds the whole json objekt
type Nbiot struct {
	Data []Data `json:"data"`
}

// NDataDecoded holds the N last DataDecoded

// GetAllRawData returns a Nbiot struct with the whole json object.
func GetAllRawData(sensorID string, username string, passwd string) (Nbiot, error) {
	var url = fmt.Sprintf("https://in.nbiot.engineering/devices/%s/data", sensorID)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(username, passwd)
	resp, err := client.Do(req)
	if err != nil {
		var null Nbiot
		return null, err
	}

	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var d Nbiot
	for {
		if err := decoder.Decode(&d); err == io.EOF {
			break
		} else if err != nil {
			return d, err
		}
	}
	return d, nil
}

// GetTelenorData returns the last data in a Data struct
func GetTelenorData(sensorID string, secrets secret.Info) (dataStruct.Data, error) {
	d, err := GetAllRawData(sensorID, secrets.Username, secrets.Password)
	if err != nil {
		var null dataStruct.Data
		return null, err
	}
	var data dataStruct.Data
	data.ID = d.Data[len(d.Data)-1].Imei
	temp := d.Data[len(d.Data)-1].Timestamp / NanoToSec
	data.Time = time.Unix(int64(temp), 0)
	data.Data = string(d.Data[len(d.Data)-1].Payload)

	return data, nil
}

// GetLastNData returns the n last data in a NDataDecoded struct
func GetLastNData(url string, username string, passwd string, n int) ([]dataStruct.Data, error) {
	d, err := GetAllRawData(url, username, passwd)
	if err != nil {
		var null []dataStruct.Data
		return null, err
	}
	var datas []dataStruct.Data
	for n > 0 {
		var data dataStruct.Data
		data.ID = d.Data[len(d.Data)-1-n].Imei
		temp := d.Data[len(d.Data)-1-n].Timestamp / NanoToSec
		data.Time = time.Unix(int64(temp), 0)
		data.Data = string(d.Data[len(d.Data)-1-n].Payload)
		datas = append(datas, data)
		n--
	}
	return datas, nil
}

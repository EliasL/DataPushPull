package ubiikget

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	secret "../../secrets"
	dataStruct "../DataStruct"
)

func floatToString(inputNum float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(inputNum, 'f', 6, 64)
}

// UbiikData is the structure of the json response
type UbiikData struct {
	Data []struct {
		ServerTime    time.Time `json:"server_time"`
		Created       time.Time `json:"created"`
		Ack           bool      `json:"ack"`
		EdeviceID     string    `json:"edevice_id"`
		BasestationID string    `json:"basestation_id"`
		Data          string    `json:"data"`
		ID            int       `json:"id"`
	} `json:"data"`
}

// GetUbiikData : Pull data from sensor with id
func GetUbiikData(sensorID string, secrets secret.Info) (dataStruct.Data, error) {
	var format = "2006-01-02T15:04:05.000Z"
	var fromDateStr = time.Now().Add(-15 * time.Second).UTC().Format(format)
	var toDateStr = time.Now().UTC().Format(format)

	client := &http.Client{}
	url := "http://wpkit.ubiik.com/api/uplink/?" +
		"to_server_time=" + toDateStr +
		"&from_server_time=" + fromDateStr
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		var null dataStruct.Data
		return null, err
	}
	req.Header.Set("Authorization", "Token "+secrets.APIKey)
	res, err := client.Do(req)
	if err != nil {
		var null dataStruct.Data
		return null, err
	}
	var jsonData UbiikData
	buf := new(bytes.Buffer)
	buf.ReadFrom(res.Body)
	jsonString := buf.String()
	fmt.Println(jsonString)
	json.Unmarshal([]byte(jsonString), &jsonData)

	/*
		var d dataStruct.Data
		layout := "2006-01-02 15:04:05"
	*/
	fmt.Println(jsonData)

	var null dataStruct.Data
	return null, nil
}

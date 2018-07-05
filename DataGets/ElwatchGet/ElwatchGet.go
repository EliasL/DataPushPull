package elwatchget

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

func floatToString(inputNum float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(inputNum, 'f', 6, 64)
}

// ElwatchSensor is the structure of the json response
type ElwatchSensor struct {
	Sensors []struct {
		Status          string  `json:"status"`
		CustomRangeLow  float64 `json:"custom_range_low"`
		LastRssi        int     `json:"last_rssi"`
		TypeID          int     `json:"type_id"`
		BinaryCutoff    float64 `json:"binary_cutoff"`
		Alias           string  `json:"alias"`
		Si              string  `json:"si"`
		Valid           int     `json:"valid"`
		Sn              int     `json:"sn"`
		Dau             string  `json:"dau"`
		CustomRangeHigh float64 `json:"custom_range_high"`
		Regcode         string  `json:"regcode"`
		LastValue2      float64 `json:"last_value2"`
		LastValue       float64 `json:"last_value"`
		LastTime        string  `json:"last_time"`
		Binarytype      int     `json:"binarytype"`
	} `json:"sensors"`
}

// // AllDataDecoded holds all received data
// type AllDataDecoded struct {
// 	AllData []DataDecoded
// }

// GetElwatchData : Pull data from sensor with id
func GetElwatchData(sensorID string, APIKey string) GetData {

	client := &http.Client{}
	url := "https://neuron.el-watch.com/api/sensordata/" + sensorID
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("API-Key", APIKey)
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	var jsonData ElwatchSensor
	buf := new(bytes.Buffer)
	buf.ReadFrom(res.Body)
	jsonString := buf.String()
	json.Unmarshal([]byte(jsonString), &jsonData)

	var d Data
	layout := "2006-01-02 15:04:05"
	d.Data = floatToString(jsonData.Sensors[0].LastValue)
	d.ID = sensorID
	d.Time, _ = time.Parse(layout, string(jsonData.Sensors[0].LastTime))
	return d
}

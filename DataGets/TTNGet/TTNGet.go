package ttnget

import (
	"time"

	secret "../../secrets"
	dataStruct "../DataStruct"

	ttnsdk "github.com/TheThingsNetwork/go-app-sdk"
	"github.com/apex/log"
)

const (
	sdkClientName = "test"
)

// GetTTNData : Pull data from node with id
func GetTTNData(deviceID string, secrets secret.Info) (dataStruct.Data, error) {
	var appID = "temp_reader"
	config := ttnsdk.NewCommunityConfig(sdkClientName)

	client := config.NewClient(appID, secrets.APIKey)
	defer client.Close()

	pubsub, err := client.PubSub()
	if err != nil {
		var null dataStruct.Data
		return null, err
	}

	defer pubsub.Close()

	myNewDevicePubSub := pubsub.Device(deviceID)
	defer myNewDevicePubSub.Close()
	uplink, err := myNewDevicePubSub.SubscribeUplink()
	if err != nil {
		var null dataStruct.Data
		return null, err
	}
	var d dataStruct.Data
	layout := "2006-01-02T15:04:05Z"
	for message := range uplink {
		d.Data = string(message.PayloadRaw)
		d.ID = deviceID
		temp, _ := message.Metadata.Gateways[0].Time.MarshalText()

		loc, _ := time.LoadLocation("Europe/Oslo")
		tempTime, _ := time.ParseInLocation(layout, string(temp), loc)
		tempTime = tempTime.Add(time.Hour * time.Duration(2))
		d.Time = tempTime

		// ubsubscribe from uplink
		err = myNewDevicePubSub.UnsubscribeUplink()
		if err != nil {
			log.WithError(err).Fatalf("Could not unsubscribe from uplink")
		}
		break
	}
	return d, nil
}

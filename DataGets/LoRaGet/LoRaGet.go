package loraget

import (
	"fmt"
	"time"

	dataStruct "../../DataGets"

	ttnsdk "github.com/TheThingsNetwork/go-app-sdk"
	"github.com/apex/log"
)

// // AllDataDecoded holds all received data
// type AllDataDecoded struct {
// 	AllData []DataDecoded
// }

const (
	sdkClientName = "test"
)

func GetTTNData(appID string, appAccessKey string) dataStruct.Data {

	config := ttnsdk.NewCommunityConfig(sdkClientName)

	client := config.NewClient(appID, appAccessKey)
	defer client.Close()

	pubsub, err := client.PubSub()
	if err != nil {
		fmt.Println(err)
		log.WithError(err).Fatal("my-amazing-app: could not get application pub/sub")

	}

	defer pubsub.Close()

	myNewDevicePubSub := pubsub.Device("lora_device_1")
	defer myNewDevicePubSub.Close()
	uplink, err := myNewDevicePubSub.SubscribeUplink()
	if err != nil {
		log.WithError(err).Fatal("my-amazing-app: could not subscribe to uplink messages")
	}

	log.Debug("After this point, the program won't show anything until we receive an uplink message from device my-new-device.")

	var d dataStruct.Data
	layout := "2006-01-02T15:04:05Z"
	for message := range uplink {
		d.Data = string(message.PayloadRaw)
		d.ID = string(message.DevID)
		temp, _ := message.Metadata.Gateways[0].Time.MarshalText()
		tempTime, _ := time.Parse(layout, string(temp))
		d.Time = tempTime
		break
	}
	return d
}

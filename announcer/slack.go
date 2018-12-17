package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/nlopes/slack"
)

func printMsg(msg string) {
	fmt.Println(time.Now(), "#", msg)
}
func msgColor(event string) string {
	switch event {
	case "UNPLANNED_INCIDENT":
		return "bad"
	case "PLANNED":
		return "warning"
	case "ANNOUNCEMENT":
		return "good"
	default:
		return "#439FE0"
	}
}

type Slack struct {
	Token   string
	Channel string
}

func postMsg(vault_server string, vault_path string, vault_key string, data []byte) {

	var notifyevent notifyEvent
	if err := json.Unmarshal(data, &notifyevent); err != nil {
		panic(err)
	}

	err := InitVault(vault_server, vault_key)
	if err != nil {
		log.Println(err)

	} else {
		printMsg("Connecting to Vault")
	}
	secretValues, err := VClient.Logical().Read(vault_path)
	if err != nil {
		log.Println(err)

	}
	var sc Slack
	for k, v := range secretValues.Data {
		switch k {
		case "slack_token":
			sc.Token = fmt.Sprintf("%v", v)
		case "slack_channel":
			sc.Channel = fmt.Sprintf("%v", v)
		default:

		}
	}

	api := slack.New(sc.Token)
	attachment := slack.Attachment{
		Color: msgColor(notifyevent.Data.Keyname),
		Fields: []slack.AttachmentField{

			slack.AttachmentField{
				Title: "Event Type",
				Value: notifyevent.Data.Keyname,
			},
			slack.AttachmentField{
				Title: notifyevent.Data.Sub,
				Value: notifyevent.Data.Sum,
			},
			slack.AttachmentField{
				Title: "Ticket Id",
				Value: strconv.Itoa(notifyevent.Data.Ticketid),
			},
		},
	}

	channelID, timestamp, err := api.PostMessage(sc.Channel, slack.MsgOptionText("", false), slack.MsgOptionAttachments(attachment))
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	fmt.Println("Message successfully sent", channelID, timestamp)
}

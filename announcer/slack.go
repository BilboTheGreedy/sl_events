package main

import (
	"encoding/json"
	"fmt"
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

func postMsg(Token string, Chan string, data []byte) {

	var notifyevent notifyEvent
	if err := json.Unmarshal(data, &notifyevent); err != nil {
		panic(err)
	}
	api := slack.New(Token)
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

	channelID, timestamp, err := api.PostMessage(Chan, slack.MsgOptionText("", false), slack.MsgOptionAttachments(attachment))
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	fmt.Println("Message successfully sent", channelID, timestamp)
}

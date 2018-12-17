package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/lib/pq"
	"github.com/nlopes/slack"
)

var VClient *api.Client // global variable
const (
	host     = "db"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "postgres"
)

type notifyEvent struct {
	Table  string `json:"table"`
	Action string `json:"action"`
	Data   struct {
		ID        int    `json:"id"`
		Ticketid  int    `json:"ticketid"`
		Sum       string `json:"sum"`
		Sub       string `json:"Sub"`
		Keyname   string `json:"keyname"`
		startDate string `json:"startDate"`
		Posted    bool   `json:"posted"`
	} `json:"data"`
}

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
			slack.AttachmentField{
				Title: "start Date",
				Value: notifyevent.Data.startDate,
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
func waitForNotification(l *pq.Listener, vault_server string, vault_path string, vault_key string) {
	for {
		select {
		case n := <-l.Notify:
			fmt.Println("Received data from channel [", n.Channel, "] :")
			// Prepare notification payload for pretty print
			var prettyJSON bytes.Buffer
			err := json.Indent(&prettyJSON, []byte(n.Extra), "", "\t")
			if err != nil {
				fmt.Println("Error processing JSON: ", err)
				return
			}
			fmt.Println(string(prettyJSON.Bytes()))
			postMsg(vault_server, vault_path, vault_key, prettyJSON.Bytes())
			return
		case <-time.After(90 * time.Second):
			fmt.Println("Received no events for 90 seconds, checking connection")
			go func() {
				l.Ping()
			}()
			return
		}
	}
}

func InitVault(server string, token string) error {
	conf := &api.Config{
		Address: server,
	}

	client, err := api.NewClient(conf)
	if err != nil {
		log.Println(err)
		return err
	}
	VClient = client

	VClient.SetToken(token)
	return nil
}

func main() {
	vault_server := flag.String("vault_server", "", " vault uri.")
	vault_path := flag.String("vault_path", "", " vault path.")
	vault_key := flag.String("vault_key", "", " vault key")

	flag.Parse()

	conninfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	_, err := sql.Open("postgres", conninfo)
	if err != nil {
		panic(err)
	}

	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	listener := pq.NewListener(conninfo, 10*time.Second, time.Minute, reportProblem)
	err = listener.Listen("events")
	if err != nil {
		panic(err)
	}

	fmt.Println("Start monitoring for new events...")
	for {
		waitForNotification(listener, *vault_server, *vault_path, *vault_key)
	}
}

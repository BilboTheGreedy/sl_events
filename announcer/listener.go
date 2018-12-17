package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"time"

	"github.com/lib/pq"
)

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
		ID       int    `json:"id"`
		Ticketid int    `json:"ticketid"`
		Sum      string `json:"sum"`
		Sub      string `json:"Sub"`
		Keyname  string `json:"keyname"`
		Created  string `json:"created"`
		Posted   bool   `json:"posted"`
	} `json:"data"`
}

func waitForNotification(l *pq.Listener) {
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

			postMsg(slack_token, slack_channel, prettyJSON.Bytes())
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

func main() {
	vault_server := flag.String("vault_server", "", " vault uri.")
	vault_path := flag.String("vault_path", "", " vault path.")
	vault_key := flag.String("vault_key", "", " vault key")

	flag.Parse()
	readVault(*vault_server, *vault_path, *vault_key)

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
		waitForNotification(listener)
	}
}

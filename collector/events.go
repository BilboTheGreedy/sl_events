package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hashicorp/vault/api"
	_ "github.com/lib/pq"
	"github.com/softlayer/softlayer-go/filter"
	"github.com/softlayer/softlayer-go/services"
	"github.com/softlayer/softlayer-go/session"
)

var sl_secret string    // global variable
var sl_user string      // global variable
var VClient *api.Client // global variable

const (
	host     = "db"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "postgres"
)

func insertDB(ticketid int, sum string, sub string, keyname string, startDate string) {

	sqlStatement := `
		INSERT INTO events (ticketid, sum, sub, keyname, startDate)
		VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING`

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	_, err = db.Exec(sqlStatement, ticketid, sum, sub, keyname, startDate)
	if err != nil {
		panic(err)
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

func getEvents(vault_server string, vault_key string, vault_path string) {
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
	for k, v := range secretValues.Data {
		// Create SoftLayer API session
		sess := session.New(k, fmt.Sprintf("%v", v))
		//Some good properties to include... maybe
		mask := "notificationOccurrenceEventType"
		//how far back we wanna look
		t := time.Now().Add(time.Duration(-1) * time.Hour)
		//Filter after date
		filters := filter.Build(
			filter.Path("startDate").DateAfter(t.Format("01/02/2006 00:00:00")),
		)

		events, err := services.GetNotificationOccurrenceEventService(sess).Mask(mask).Filter(filters).GetAllObjects()
		if err != nil {
			fmt.Printf("\n Unable to get events:\n - %s\n", err)
			os.Exit(1)
		}

		for _, event := range events {
			printMsg("Adding " + *event.NotificationOccurrenceEventType.KeyName)
			insertDB(*event.SystemTicketId, *event.Summary, *event.Subject, *event.NotificationOccurrenceEventType.KeyName, fmt.Sprintf("%s", *event.StartDate))
		}

	}

}

func printMsg(msg string) {
	fmt.Println(time.Now(), "#", msg)
}

func main() {
	vault_server := flag.String("vault_server", "", " vault uri.")
	vault_path := flag.String("vault_path", "", " vault path.")
	vault_key := flag.String("vault_key", "", " vault key")
	flag.Parse()

	//time between polls
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				//Call the periodic function.
				getEvents(*vault_server, *vault_key, *vault_path)
			}
		}
	}()

	quit := make(chan bool, 1)
	// main will continue to wait untill there is an entry in quit
	<-quit
}

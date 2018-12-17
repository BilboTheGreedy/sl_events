package main

import (
	"fmt"
	"log"

	"github.com/hashicorp/vault/api"
)

var VClient *api.Client  // global variable
var slack_channel string // global variable
var slack_token string   // global variable

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

func readVault(server string, keyName string, token string) {
	err := InitVault(server, token)
	if err != nil {
		log.Println(err)

	} else {
		printMsg("Connecting to Vault")
	}
	secretValues, err := VClient.Logical().Read(keyName)
	if err != nil {
		log.Println(err)

	}
	//TODO: make sure to select correct kv pair if multiple in data
	for k, v := range secretValues.Data {
		secretSwitch(k, fmt.Sprintf("%v", v), keyName)
	}

}

func secretSwitch(k string, v string, keyName string) {
	switch k {
	case "slack_token":
		slack_token = v
	case "slack_channel":
		slack_channel = v
	default:

	}

}

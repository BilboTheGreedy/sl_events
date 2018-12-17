package main

import (
	"log"

	"github.com/hashicorp/vault/api"
)

var VClient *api.Client // global variable

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

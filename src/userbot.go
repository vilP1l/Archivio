package main

import (
	"fmt"
	"log"
	"github.com/bwmarrin/discordgo"
	"github.com/logrusorgru/aurora"
)

var userSession *discordgo.Session

func connectUser(token string) {
	client, err := discordgo.New(token)

	if err != nil {
		log.Fatalln("Failed to login to discord:", err)
	}

	err = client.Open()

	if err != nil {
		log.Fatalln("Failed to open connection to discord:", err)
	}
	
    fmt.Println(aurora.Green("User bot logged in!"))

    client.UserAgent = ""
    
	userSession = client
}
package main

import (
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
)

var DISCORD_TOKEN = os.Getenv("DISCORD_TOKEN")
var DISCORD_CHANNEL = os.Getenv("DISCORD_CHANNEL")

// Variables used for command line parameters
type DiscordBot struct {
	Session  *discordgo.Session
	Token    string
	Notified map[int32]bool
	Startup  chan string
}

func (discord *DiscordBot) NewDiscordBot() {

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + discord.Token)
	if err != nil {
		log.Println("error creating Discord session,", err)
		return
	}

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.IntentsGuildMessages
	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		log.Println("error opening connection,", err)
		return
	}
	discord.Session = dg
	discord.Startup <- "[DEBUG] Discord bot running"
}

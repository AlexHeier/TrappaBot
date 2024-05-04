package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"trappabot/functions"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var channelReview = "1013473566806786058"

func main() {

	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			fmt.Println("Warning: Error loading .env file:", err)
		}
	} else if os.IsNotExist(err) {
		fmt.Println(".env file does not exist. Environment variables will be loaded from the system environment.")
	} else {
		fmt.Println("Error checking .env file:", err)
	}

	Token := os.Getenv("DISCORD_BOT_TOKEN")
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dg.AddHandler(functions.MessageReactionAdd)
	dg.AddHandler(functions.MessageReactionRemove)
	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.ChannelID == channelReview { // if its the channel for stair review
		if m.Attachments[0].Height == 0 {
			functions.ImageToOpenAI(s, m) // if it is a image
		}
	}

}

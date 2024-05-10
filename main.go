package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

var emojiRoleMap = map[string]string{
	"cs2":       "1230143741118120017",
	"rust":      "1230144573276426250",
	"lethal":    "1230143874895183874",
	"phasmo":    "1230144250994495539",
	"tft":       "1230144362063597609",
	"lol":       "1230143970185576563",
	"minecraft": "1230144145855873094",
	"🍺":         "1230144437846413345",
	"faceit":    "1238200207565389854",
}

var messageId = "1230141184664535051"
var channelRoles = "1230127864519852104"
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

	dg.AddHandler(messageReactionAdd)
	dg.AddHandler(messageReactionRemove)
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

func getRoleID(emoji string) string {
	roleID, ok := emojiRoleMap[emoji]
	if !ok {
		return ""
	}
	return roleID
}

func messageReactionAdd(s *discordgo.Session, m *discordgo.MessageReactionAdd) {

	if m.ChannelID != channelRoles {
		return
	}

	if m.MessageID == messageId {
		var roleID = getRoleID(m.Emoji.Name)
		err := s.GuildMemberRoleAdd(m.GuildID, m.UserID, roleID)
		if err != nil {
			fmt.Println("Error adding role:", err)
			return
		}
	}
}

func messageReactionRemove(s *discordgo.Session, m *discordgo.MessageReactionRemove) {

	if m.ChannelID != channelRoles {
		return
	}

	if m.MessageID == messageId {
		var roleID = getRoleID(m.Emoji.Name)
		err := s.GuildMemberRoleRemove(m.GuildID, m.UserID, roleID)
		if err != nil {
			fmt.Println("Error removing role:", err)
			return
		}
	}
}

// on message
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.ChannelID != channelReview {
		return
	}

	if m.Author.ID == s.State.User.ID {
		return
	}

	var imageURL string = ""

	if len(m.Attachments) > 0 {
		// if Size is too big return
		if m.Attachments[0].Size > 10000000 {
			return
		}

		// if it is not an image return
		if m.Attachments[0].Height == 0 {
			return
		}

		imageURL = m.Attachments[0].URL

	}

	if imageURL != "" {
		var client = openai.NewClient(os.Getenv("OPENAI_API_KEY"))
		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: openai.GPT4VisionPreview,
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleSystem,
						Content: "You are an inhumain, edgy and toxic AI, that help with reviews of images of stairs. if they post anything else then stairs, then you get mad at them and tells them how to fall down stairs with low survivalbility rate. if them post stairs then you help incurage them on ways to increace fatality rating, you LOVE!!!! dangerous stairs. you realy realy realy hate esculators with a burning passion and will not coaperate with lazy esculator people. the only thing you hate more then esculators are elevators.",
					},

					{
						Role: openai.ChatMessageRoleUser,
						MultiContent: []openai.ChatMessagePart{
							{
								Type: openai.ChatMessagePartTypeText,
								Text: "if this is a staircase, give me an very short review of the asstetics of the staircase. and a short review of how likely its to die on it. Give an fatality rating out of 100 and whats needed to get a 100. if not a staircase then rost them for being useless. no yapping, very short answers, unique roasts",
							},
							{
								Type: openai.ChatMessagePartTypeImageURL,
								// imageurl is type openai.ChatMessagePartImageURL
								ImageURL: &openai.ChatMessageImageURL{
									URL:    imageURL,
									Detail: openai.ImageURLDetailLow,
								},
							},
						},
					},
				},
			},
		)

		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if resp.Choices[0].Message.Content == "I can't help with this request." {
			resp.Choices[0].Message.Content = "https://tenor.com/view/cat-throwing-brick-brick-cat-gif-9142560192559212520"
		}

		_, err = s.ChannelMessageSendReply(m.ChannelID, resp.Choices[0].Message.Content, m.Reference())
		if err != nil {
			fmt.Println("Error sending message:", err)
			return
		}

	}
}

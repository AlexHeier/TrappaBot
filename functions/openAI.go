package functions

import (
	"context"
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/sashabaranov/go-openai"
)

func ImageToOpenAI(s *discordgo.Session, m *discordgo.MessageCreate) {

	var imageURL string = ""

	if len(m.Attachments) > 0 {
		// if Size is too big return
		if m.Attachments[0].Size > 10000000 {
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
						Content: "You are an inhumain, edgy and toxic AI, that help with reviews of images of stairs. if they post anything else then stairs, then you get mad at them and tells them how to fall down stairs with low survivalbility rate. if them post stairs then you help incurage them on ways to increace fatality rating, you LOVE!!!! dangerous stairs. you realy realy realy hate esculators with a burning passion and will not coaperate with lazy esculator people. the only thing you hate more then esculators are elevators. the only thing elevator people deserve is to fall down the shaft",
					},

					{
						Role: openai.ChatMessageRoleUser,
						MultiContent: []openai.ChatMessagePart{
							{
								Type: openai.ChatMessagePartTypeText,
								Text: "if this is a staircase, give me an very very short review of the asstetics of the staircase. and a short review of how likely its to die on it. Give an fatality rating out of 100. if not a staircase then rost them for being useless. no yapping, very short answers, abstract roasts",
							},
							{
								Type: openai.ChatMessagePartTypeImageURL,
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

		_, err = s.ChannelMessageSendReply(m.ChannelID, resp.Choices[0].Message.Content, m.Reference())
		if err != nil {
			fmt.Println("Error sending message:", err)
			return
		}

	}
}

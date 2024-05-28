package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gocolly/colly"
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
	"beermug":   "1230144437846413345",
	"faceit":    "1238200207565389854",
	"SoTboat":   "1240694581654323230",
}

var TriggerChannels map[string]string = map[string]string{
	"1237466111449108480": "Faceit",
	"1236031688346177657": "CS2",
	"1236035830439743508": "Rust",
	"1240694963998687253": "SoT",
	"1236032442754797658": "League",
	"1236035330021392384": "Phasmo",
	"1236033935608385654": "Lethal",
	"1236034392481206403": "TFT",
	"1236034687592566834": "Minecraft",
}

var activeVoiceUsers = make(map[string]int)

var messageId = "1230141184664535051"
var channelRoles = "1230127864519852104"
var channelReview = "1013473566806786058"
var GuildID = "1012016741238448278"

var mood = []string{
	"toxic", "edgy", "romantic", "horny", "crazy", "mad", "loving", "pathetic", "sad", "insecure",
	"dominant", "submissive", "chad", "introverted", "extroverted", "confident", "anxious", "eccentric",
	"energetic", "lazy", "adventurous", "cautious", "ambitious", "content", "frustrated", "calm", "moody",
	"optimistic", "pessimistic", "charming", "awkward", "mysterious", "playful", "serious", "goofy", "stoic",
	"silly", "sensitive", "rebellious", "compliant", "bold", "shy", "self-assured", "self-conscious", "quirky",
	"flirtatious", "reserved", "outgoing", "withdrawn", "perfectionist", "impatient", "patient", "compassionate",
	"aloof", "attentive", "indifferent", "friendly", "distant", "passionate", "detached", "gregarious",
	"talkative", "quiet", "carefree", "selfish", "selfless", "impulsive", "logical", "emotional", "rational",
	"irrational", "stubborn", "flexible", "independent", "dependent", "practical", "dreamy", "realistic",
	"idealistic", "cynical", "easygoing", "uptight", "relaxed", "tense", "joyful", "melancholic", "nostalgic",
	"curious", "opinionated", "open-minded", "skeptical", "trustful", "confused", "focused", "scatterbrained",
	"assertive", "passive", "reclusive", "social", "inclusive", "exclusive", "outspoken", "impulsive",
	"deliberate", "spontaneous", "predictable", "loyal", "disloyal", "observant", "unobservant", "helpful",
	"unhelpful", "forgiving", "resentful", "generous", "gregarious", "adaptable", "rigid", "creative", "analytical",
	"insightful", "shallow", "deep", "indifferent", "empathetic", "apathetic", "lethargic", "unmotivated",
	"cheerful", "gloomy", "positive", "negative", "charismatic", "uncharismatic", "motivated", "uninspired",
	"determined", "resigned", "happy", "unhappy", "spontaneous", "planned",
}
var adjective = []string{
	"stupid", "idiotic", "silly", "foolish", "absurd",
	"useless", "ridiculous", "ludicrous", "pathetic",
	"clumsy", "obnoxious", "petty", "vile", "disgraceful",
	"inept", "ignorant", "incompetent", "insignificant",
	"mediocre", "nonsensical", "pointless", "senseless",
	"trivial", "worthless", "laughable", "dismal",
	"lame", "miserable", "derisory", "detestable",
}

const lastCommitFile = "lastCommit.txt"

type Commit struct {
	ID       string
	Author   string
	Message  string
	ImageURL string
}

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

	fetchCommits(dg)
	scrapeSteamStore(dg)

	dg.AddHandler(messageReactionAdd)
	dg.AddHandler(messageReactionRemove)
	dg.AddHandler(messageCreate)
	dg.AddHandler(messageHandler)
	dg.AddHandler(voiceStateUpdate)
	dg.AddHandler(updateVoiceState)

	go func() {
		for range time.Tick(2 * time.Minute) {
			fetchCommits(dg)
		}
	}()
	go func() {
		for range time.Tick(12 * time.Hour) {
			scrapeSteamStore(dg)
		}
	}()

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

		var theContext string = fmt.Sprintf("If you cannt assist with my request anser with:'I'm sorry, I can't assist with that request.' else you are a %v AI that helps review stair cases. You love staircases, do not like escelators and hate elevators. If they post anything else then a staircase you get mad at them and incurage them on ways to accidently fall down a staircase or to do it from something else.", random(mood))
		var thePromt string = fmt.Sprintf("if this is a staircase, give me an very short review of the asstetics of the staircase. and a short review of how likely its to die on it. Give an fatality rating out of 100. if not a staircase then rost them and the content of the image for being %v. no yapping, very short answers, %v roasts", random(adjective), random(mood))

		var client = openai.NewClient(os.Getenv("OPENAI_API_KEY"))
		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: openai.GPT4VisionPreview,
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleSystem,
						Content: theContext,
					},

					{
						Role: openai.ChatMessageRoleUser,
						MultiContent: []openai.ChatMessagePart{
							{
								Type: openai.ChatMessagePartTypeText,
								Text: thePromt,
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

		if resp.Choices[0].Message.Content == "I'm sorry, I can't assist with that request." {
			resp.Choices[0].Message.Content = "https://tenor.com/view/cat-throwing-brick-brick-cat-gif-9142560192559212520"
		}

		_, err = s.ChannelMessageSendReply(m.ChannelID, resp.Choices[0].Message.Content, m.Reference())
		if err != nil {
			fmt.Println("Error sending message:", err)
			return
		}

	}
}

func random(array []string) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return array[r.Intn(len(array))]
}

func fetchCommits(s *discordgo.Session) {
	c := colly.NewCollector()
	var commits []Commit

	var lastCommitID string
	if file, err := os.ReadFile(lastCommitFile); err == nil {
		lastCommitID = strings.TrimSpace(string(file))
	} else if !os.IsNotExist(err) {
		return
	}

	c.OnHTML(".commit.columns", func(e *colly.HTMLElement) {
		commitID := e.Attr("like-id")
		if lastCommitID == "" || commitID > lastCommitID {
			commit := Commit{
				ID:       commitID,
				Author:   e.ChildText(".author"),
				Message:  e.ChildText(".commits-message"),
				ImageURL: e.ChildAttr(".avatar img", "src"),
			}
			commits = append(commits, commit)
		}
	})

	c.OnScraped(func(_ *colly.Response) {

		sort.Slice(commits, func(i, j int) bool {
			return commits[i].ID < commits[j].ID
		})

		for _, commit := range commits {

			embed := &discordgo.MessageEmbed{
				Author: &discordgo.MessageEmbedAuthor{
					Name:    fmt.Sprintf("%s: %s", commit.Author, commit.ID),
					IconURL: commit.ImageURL,
				},
				Description: fmt.Sprintf("**Message:** %s", commit.Message),
				Color:       12648462,
				Footer: &discordgo.MessageEmbedFooter{
					Text: "Source: https://commits.facepunch.com/r/rust_reboot",
				},
				Timestamp: time.Now().Format(time.RFC3339),
			}

			_, err := s.ChannelMessageSendEmbed("1243953919621861428", embed)
			if err != nil {
				fmt.Println("Error sending embed message:", err)
			}
		}

		if len(commits) > 0 {
			lastCommit := commits[len(commits)-1]
			err := os.WriteFile(lastCommitFile, []byte(lastCommit.ID), 0644)
			if err != nil {
				return
			}
		}
	})

	c.Visit("https://commits.facepunch.com/r/rust_reboot")
}

func scrapeSteamStore(discord *discordgo.Session) {
	c := colly.NewCollector(
		colly.AllowedDomains("store.steampowered.com"),
	)

	var items []*discordgo.MessageEmbed
	var firstLink string

	c.OnHTML("#ItemDefsRows .item_def_grid_item", func(e *colly.HTMLElement) {
		itemPrice := e.ChildText(".item_def_price")
		itemName := e.ChildText(".item_def_name")
		itemImageURL := e.ChildAttr("img.item_def_icon", "src")
		itemLink := e.ChildAttr("a", "href")

		if firstLink == "" {
			firstLink = itemLink
		}

		embed := &discordgo.MessageEmbed{
			Title:       itemName,
			Description: fmt.Sprintf("Price: %s", itemPrice),
			URL:         itemLink,
			Image:       &discordgo.MessageEmbedImage{URL: itemImageURL},
			Color:       12648462,
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Item on sale from: ",
			},
			Timestamp: time.Now().Format(time.RFC3339),
		}
		items = append(items, embed)
	})

	c.OnScraped(func(_ *colly.Response) {
		prevLinkBytes, err := os.ReadFile("firstLink.txt")
		if err == nil && string(prevLinkBytes) == firstLink {
			return
		}

		err = os.WriteFile("firstLink.txt", []byte(firstLink), 0644)
		if err != nil {
			log.Println("Error writing to file:", err)
		}

		for i := 0; i < len(items); i += 10 {
			end := i + 10
			if end > len(items) {
				end = len(items)
			}

			_, err := discord.ChannelMessageSendEmbeds("1243598497916256380", items[i:end])
			if err != nil {
				log.Println("Error sending embeds:", err)
			}
		}
	})

	c.OnError(func(_ *colly.Response, err error) {
		log.Println("Something went wrong:", err)
	})

	c.Visit("https://store.steampowered.com/itemstore/252490/browse/?filter=Limited")
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.HasPrefix(m.Content, "!roles") && m.Author.ID == "259789771260428288" {
		start := strings.Index(m.Content, "(")
		end := strings.LastIndex(m.Content, ")")

		if start == -1 || end == -1 || end <= start {
			s.ChannelMessageSend(m.ChannelID, "Invalid command usage. Please enclose the message content in parentheses.")
			return
		}

		newContent := m.Content[start+1 : end]
		messageID := "1230141184664535051"
		channelID := "1230127864519852104" // Hardcoded channel ID where the message resides

		_, err := s.ChannelMessageEdit(channelID, messageID, newContent)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Failed to edit message: "+err.Error())
			return
		}

		s.ChannelMessageSend(m.ChannelID, "Message updated successfully!")
	}
}

func voiceStateUpdate(s *discordgo.Session, v *discordgo.VoiceStateUpdate) {
	var gameCategories = map[string]string{
		"Faceit":    "1238426908342091788",
		"CS2":       "1236031485694181438",
		"Rust":      "1236035753285521439",
		"SoT":       "1240694798755561582",
		"League":    "1236032359363510384",
		"Phasmo":    "1236034853053403187",
		"Lethal":    "1236033848458870846",
		"TFT":       "1236034325649035386",
		"Minecraft": "1236034481270292585",
	}

	if gameName, ok := TriggerChannels[v.ChannelID]; ok {
		channels, _ := s.GuildChannels(GuildID)
		var existingNumbers []int
		prefix := fmt.Sprintf("\U0001F509 | %s Voice ", gameName) // Unicode voice emoji

		for _, channel := range channels {
			if strings.HasPrefix(channel.Name, prefix) {
				numberStr := strings.TrimPrefix(channel.Name, prefix)
				if num, err := strconv.Atoi(numberStr); err == nil {
					existingNumbers = append(existingNumbers, num)
				}
			}
		}

		sort.Ints(existingNumbers)
		newNumber := 1
		for _, num := range existingNumbers {
			if num == newNumber {
				newNumber++
			} else {
				break
			}
		}

		newChannelName := fmt.Sprintf("%s%d", prefix, newNumber)
		parentID, exists := gameCategories[gameName]
		if !exists {
			fmt.Println("Category ID not found for the game:", gameName)
			return
		}

		channel, err := s.GuildChannelCreateComplex(GuildID, discordgo.GuildChannelCreateData{
			Name:     newChannelName,
			Type:     discordgo.ChannelTypeGuildVoice,
			ParentID: parentID,
		})
		if err != nil {
			fmt.Println("Error creating channel:", err)
			return
		}

		err = s.GuildMemberMove(GuildID, v.UserID, &channel.ID)
		if err != nil {
			fmt.Println("Error moving member:", err)
			return
		}

		activeVoiceUsers[channel.ID] = 1

		go func() {
			for {
				<-time.After(3 * time.Second)
				if activeVoiceUsers[channel.ID] == 0 {
					s.ChannelDelete(channel.ID)
					break
				}
			}
		}()
	}
}

func updateVoiceState(s *discordgo.Session, v *discordgo.VoiceStateUpdate) {
	if v.BeforeUpdate != nil && v.BeforeUpdate.ChannelID != "" {
		activeVoiceUsers[v.BeforeUpdate.ChannelID]--
	}
	if v.ChannelID != "" {
		activeVoiceUsers[v.ChannelID]++
	}
}

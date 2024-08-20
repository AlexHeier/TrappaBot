package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/bwmarrin/discordgo"     // Discord handling
	"github.com/gocolly/colly"          // Web scraping
	"github.com/jackc/pgx/v4/pgxpool"   // Database connection
	"github.com/joho/godotenv"          // env variables
	"github.com/sashabaranov/go-openai" // OpenAI
)

// Variables to manage command-line flags.
var GuildIDflag = flag.String("guild", "", "Test guild ID. If not passed - bot registers commands globally")
var RemoveCommands = flag.Bool("rmcmd", true, "Remove all commands after shutdowning or not")

// Global variables for managing Discord session and database pool.
var dg *discordgo.Session
var dbpool *pgxpool.Pool

// Variables to identify specific channels and messages.
var messageId = "1230141184664535051"
var channelRoles = "1230127864519852104"
var channelReview = "1013473566806786058"
var GuildID = "1012016741238448278"
var rustUpdateChannel = "1243953919621861428"

// Constant for managing file paths.
const lastCommitFile = "lastCommit.txt"

// Struct for holding commit data.
type Commit struct {
	ID       string
	Author   string
	Message  string
	ImageURL string
}

// Arrays holding different mood and adjective descriptions for openAI prompts and respose.
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
var gifs = []string{"https://tenor.com/view/shut-up-shush-shh-ok-bird-gif-17679708", "https://tenor.com/view/cat-throwing-brick-brick-cat-gif-9142560192559212520",
	"https://tenor.com/view/bonk-gif-26414884", "https://tenor.com/view/byuntear-sad-sad-cat-cat-meme-gif-12058012318069999477",
	"https://tenor.com/view/cat-cat-fight-annoyed-annoyed-cat-gif-2039169220993261012", "https://tenor.com/view/donald-sutherland-goddamn-disappointment-don-donald-mechanic-gif-23830461",
	"https://tenor.com/view/why-michael-scott-the-office-why-are-the-way-that-you-are-gif-5593972", "https://tenor.com/view/its-always-sunny-dennis-reynolds-dumb-driving-you-dumb-bitch-gif-16430416",
	"https://tenor.com/view/robin-williams-mrs-doubtfire-get-lost-gif-16430428", "https://tenor.com/view/obama-confused-why-why-tho-but-why-gif-16823464",
	"https://tenor.com/view/memes-2022-funny-dirty-gif-26230406", "https://tenor.com/view/whyareyougay-uganda-gay-gif-14399349",
	"https://tenor.com/view/down-syndrome-huh-look-back-what-wtf-gif-14728372", "https://tenor.com/view/you-gif-25833251",
}

// Definitions for Discord application commands.
var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "creategame",
		Description: "creates everything for adding a new game to the discord",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "game_name",
				Description: "the name of the game to be created",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "name_abbreviation",
				Description: "short version of the name to be used for the channel",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "emoji",
				Description: "The emoji auto complited to image",
				Required:    true,
			},
		},
	},
	{
		Name:        "deletegame",
		Description: "deletes a game category and all things related.",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "game_name",
				Description: "the name of the game to be deleted",
				Required:    true,
			},
		},
	},
}

// commandHandlers maps command names to their respective handler functions.
var commandHandlers = map[string]func(dg *discordgo.Session, i *discordgo.InteractionCreate){

	// createGame handles the creation of a new game category, role, and associated channels in Discord.
	// It inserts related information into the database and sets necessary permissions.
	// @param dg the Discord session
	// @param i the interaction creation event that triggered this command
	"creategame": func(dg *discordgo.Session, i *discordgo.InteractionCreate) {
		if err := acknowledgeInteraction(dg, i); err != nil {
			return
		}

		adminUserID := os.Getenv("ADMIN_ID")
		if i.Member.User.ID != adminUserID {
			response := "You do not have permission to use this command."
			if err := sendResponse(dg, i, response); err != nil {
				log.Printf("Error sending response: %v", err)
			}
			return
		}

		gameName := strings.ToLower(i.ApplicationCommandData().Options[0].StringValue())
		abbreviation := i.ApplicationCommandData().Options[1].StringValue()
		emoji := i.ApplicationCommandData().Options[2].StringValue()
		randomColor := rand.Intn(0xFFFFFF + 1)
		hoist := true
		mentionable := true

		newRole, err := dg.GuildRoleCreate(GuildID, &discordgo.RoleParams{
			Name:        abbreviation,
			Color:       &randomColor,
			Hoist:       &hoist,
			Mentionable: &mentionable,
		})
		if err != nil {
			response := fmt.Sprintf("Failed to create role %s", abbreviation)
			if err := sendResponse(dg, i, response); err != nil {
				log.Printf("Error sending response: %v", err)
			}
			return
		}

		permissionOverwrites := []*discordgo.PermissionOverwrite{
			{
				ID:    newRole.ID,
				Type:  discordgo.PermissionOverwriteTypeRole,
				Allow: 0x00000400 | 0x00100000 | 0x00200000 | 0x00000800, // ViewChannel | Connect | Speak | SendMessages
			},
			{
				ID:   GuildID,
				Type: discordgo.PermissionOverwriteTypeRole,
				Deny: 0x00000400, // ViewChannel
			},
		}

		category, err := dg.GuildChannelCreateComplex(GuildID, discordgo.GuildChannelCreateData{
			Name:                 "--------- | " + gameName + " | ---------",
			Type:                 discordgo.ChannelTypeGuildCategory,
			PermissionOverwrites: permissionOverwrites,
		})
		if err != nil {
			response := "Error creating category."
			if err := sendResponse(dg, i, response); err != nil {
				log.Printf("Error sending response: %v", err)
			}
			return
		}

		_, err = dg.GuildChannelCreateComplex(GuildID, discordgo.GuildChannelCreateData{
			Name:                 "\U0001F4C4|" + abbreviation + "-chat",
			Type:                 discordgo.ChannelTypeGuildText,
			ParentID:             category.ID,
			PermissionOverwrites: permissionOverwrites,
		})
		if err != nil {
			response := "Error creating text channel."
			if err := sendResponse(dg, i, response); err != nil {
				log.Printf("Error sending response: %v", err)
			}
			return
		}

		voiceChannel, err := dg.GuildChannelCreateComplex(GuildID, discordgo.GuildChannelCreateData{
			Name:                 "\U0001F50A | " + abbreviation,
			Type:                 discordgo.ChannelTypeGuildVoice,
			ParentID:             category.ID,
			PermissionOverwrites: permissionOverwrites,
		})
		if err != nil {
			response := "Error creating voice channel."
			if err := sendResponse(dg, i, response); err != nil {
				log.Printf("Error sending response: %v", err)
			}
			return
		}

		// Insert data into the database
		tx, err := dbpool.Begin(context.Background())
		if err != nil {
			log.Println("Failed to start transaction:", err)
			return
		}

		defer tx.Rollback(context.Background())

		_, err = tx.Exec(context.Background(), "INSERT INTO chategory (chategoryID, roleID, name, emoji, abbreviation) VALUES ($1, $2, $3, $4, $5)", category.ID, newRole.ID, gameName, emoji, abbreviation)
		if err != nil {
			log.Println("Failed to insert into chategory:", err)
			return
		}

		_, err = tx.Exec(context.Background(), "INSERT INTO mainVoice (chategoryID, channelID) VALUES ($1, $2)", category.ID, voiceChannel.ID)
		if err != nil {
			log.Println("Failed to insert into mainVoice:", err)
			return
		}

		if err := tx.Commit(context.Background()); err != nil {
			log.Println("Failed to commit transaction:", err)
			return
		}

		// Update permissions for the specified categories
		specifiedCategories := []string{"1012016741850820698", "1012016741850820699"}
		for _, catID := range specifiedCategories {
			err = dg.ChannelPermissionSet(catID, newRole.ID, discordgo.PermissionOverwriteTypeRole, 0x00000400|0x00100000|0x00200000|0x00000800, 0)
			if err != nil {
				log.Printf("Error setting permissions for category %s: %v", catID, err)
			}
		}

		if err := updateRoleMessage(dg); err != nil {
			response := fmt.Sprintf("Error updating role message: %v", err)
			if err := sendResponse(dg, i, response); err != nil {
				log.Printf("Error sending response: %v", err)
			}
		}

		response := fmt.Sprintf("Game %s created successfully", gameName)
		if err := sendResponse(dg, i, response); err != nil {
			log.Printf("Error sending response: %v", err)
		}
	},

	// deleteGame handles the deletion of a game category and all associated channels and roles in Discord.
	// It also removes corresponding entries from the database.
	// @param dg the Discord session
	// @param i the interaction creation event that triggered this command
	"deletegame": func(dg *discordgo.Session, i *discordgo.InteractionCreate) {
		if err := acknowledgeInteraction(dg, i); err != nil {
			return
		}

		adminUserID := os.Getenv("ADMIN_ID")
		if i.Member.User.ID != adminUserID {
			response := "You do not have permission to use this command."
			if err := sendResponse(dg, i, response); err != nil {
				log.Printf("Error sending response: %v", err)
			}
			return
		}

		name := strings.ToLower(i.ApplicationCommandData().Options[0].StringValue())

		// Start a database transaction
		tx, err := dbpool.Begin(context.Background())
		if err != nil {
			response := fmt.Sprintf("Failed to start database transaction. Err: %s", err)
			if err := sendResponse(dg, i, response); err != nil {
				log.Printf("Error sending response: %v", err)
			}
			return
		}
		defer tx.Rollback(context.Background())

		var chategoryID, roleID string
		err = tx.QueryRow(context.Background(), "SELECT chategoryID, roleID FROM chategory WHERE name = $1", name).Scan(&chategoryID, &roleID)
		if err != nil {
			response := fmt.Sprintf("Failed to find game with the given name. Err: %s", err)
			if err := sendResponse(dg, i, response); err != nil {
				log.Printf("Error sending response: %v", err)
			}
			return
		}

		// Fetch and delete all child channels
		channels, err := dg.GuildChannels(GuildID)
		if err != nil {
			response := fmt.Sprintf("Failed to fetch channels. Err: %s", err)
			if err := sendResponse(dg, i, response); err != nil {
				log.Printf("Error sending response: %v", err)
			}
			return
		}
		for _, channel := range channels {
			if channel.ParentID == chategoryID {
				if _, err = dg.ChannelDelete(channel.ID); err != nil {
					response := fmt.Sprintf("Failed to delete child channel in Discord. Err: %s", err)
					if err := sendResponse(dg, i, response); err != nil {
						log.Printf("Error sending response: %v", err)
					}
					return
				}
			}
		}

		// Delete the category
		_, err = dg.ChannelDelete(chategoryID)
		if err != nil {
			response := fmt.Sprintf("Failed to delete category channel in Discord. Err: %s", err)
			if err := sendResponse(dg, i, response); err != nil {
				log.Printf("Error sending response: %v", err)
			}
			return
		}

		// Delete the role in Discord
		err = dg.GuildRoleDelete(GuildID, roleID)
		if err != nil {
			response := fmt.Sprintf("Failed to delete role in Discord. Err: %s", err)
			if err := sendResponse(dg, i, response); err != nil {
				log.Printf("Error sending response: %v", err)
			}
			return
		}

		_, err = tx.Exec(context.Background(), "DELETE FROM chategory WHERE chategoryID = $1", chategoryID)
		if err != nil {
			response := fmt.Sprintf("Failed to delete entry from chategory table. Err: %s", err)
			if err := sendResponse(dg, i, response); err != nil {
				log.Printf("Error sending response: %v", err)
			}
			return
		}

		// Commit the transaction
		if err := tx.Commit(context.Background()); err != nil {
			response := fmt.Sprintf("Failed to commit database transaction. Err: %s", err)
			if err := sendResponse(dg, i, response); err != nil {
				log.Printf("Error sending response: %v", err)
			}
			return
		}

		if err := updateRoleMessage(dg); err != nil {
			response := fmt.Sprintf("Error updating role message: %v", err)
			if err := sendResponse(dg, i, response); err != nil {
				log.Printf("Error sending response: %v", err)
			}
		}

		// Send a success response
		response := "Game category, all child channels, and role successfully deleted."
		if err := sendResponse(dg, i, response); err != nil {
			log.Printf("Error sending response: %v", err)
		}
	},
}

// init initializes command line flags and loads environment variables.
func init() {
	flag.Parse()
	envErr := godotenv.Load()
	if envErr != nil {
		log.Fatalf("Error loading .env file: %v", envErr)
	}

	BotToken := os.Getenv("DISCORD_BOT_TOKEN")

	var err error
	dg, err = discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}

	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}

// connectToDB establishes a connection to the PostgreSQL database.
func connectToDB() {

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	var err error
	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUser, dbPassword, dbHost, dbPort, dbName)
	dbpool, err = pgxpool.Connect(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	fmt.Println("Connected to database.")

}

// main sets up the Discord bot connection and command handling.
func main() {

	connectToDB()
	defer dbpool.Close()

	fetchCommits(dg)
	scrapeSteamStore(dg)

	dg.AddHandler(messageReactionAdd)
	dg.AddHandler(messageReactionRemove)
	dg.AddHandler(messageCreate)
	dg.AddHandler(voiceStateUpdate)
	dg.AddHandler(updateVoiceState)
	dg.AddHandler(messageTimer)
	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	go func() {
		for range time.Tick(2 * time.Minute) {
			fetchCommits(dg)
		}
	}()
	go func() {
		for range time.Tick(6 * time.Hour) {
			scrapeSteamStore(dg)
		}
	}()

	err := dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := dg.ApplicationCommandCreate(dg.State.User.ID, *GuildIDflag, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	defer dg.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	if *RemoveCommands {
		log.Println("Removing commands...")

		registeredCommands, err := dg.ApplicationCommands(dg.State.User.ID, *GuildIDflag)
		if err != nil {
			log.Fatalf("Could not fetch registered commands: %v", err)
		}

		for _, v := range registeredCommands {
			err := dg.ApplicationCommandDelete(dg.State.User.ID, *GuildIDflag, v.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
	}
}

// getRoleID retrieves the role ID based on emoji data.
// @param emoji the emoji to lookup
// @return string the corresponding role ID
func getRoleID(emoji string) string {
	var roleID string

	// Check if the emoji is a custom emoji (only digits)
	isCustomEmoji := true
	for _, char := range emoji {
		if !unicode.IsDigit(char) {
			isCustomEmoji = false
			break
		}
	}

	if isCustomEmoji {
		// Query the entire database for custom emoji
		rows, err := dbpool.Query(context.Background(), "SELECT emoji, roleID FROM chategory")
		if err != nil {
			fmt.Println("Error fetching emojis and roles:", err)
			return ""
		}
		defer rows.Close()

		for rows.Next() {
			var dbEmoji, dbRoleID string
			if err := rows.Scan(&dbEmoji, &dbRoleID); err != nil {
				fmt.Println("Error scanning row:", err)
				continue
			}

			// Check if the database emoji contains the emoji ID
			if strings.Contains(dbEmoji, emoji) {
				return dbRoleID
			}
		}

		if err := rows.Err(); err != nil {
			fmt.Println("Error with rows:", err)
		}
		return ""
	}

	// For standard emojis, query directly
	err := dbpool.QueryRow(context.Background(), "SELECT roleID FROM chategory WHERE emoji = $1", emoji).Scan(&roleID)
	if err != nil {
		fmt.Println("Error getting role ID:", err)
		return ""
	}
	return roleID
}

// messageReactionAdd handles adding roles when a reaction is added to a message.
// @param s the Discord session
// @param m the reaction addition event
func messageReactionAdd(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	if m.ChannelID != channelRoles {
		return
	}

	if m.MessageID == messageId {

		// Determine the emoji APIName to use
		var roleID string
		if m.Emoji.ID != "" {
			// Custom emoji
			roleID = getRoleID(m.Emoji.ID)
		} else {
			// Standard emoji
			roleID = getRoleID(m.Emoji.Name)
		}

		if roleID == "" {
			return
		}
		err := s.GuildMemberRoleAdd(m.GuildID, m.UserID, roleID)
		if err != nil {
			fmt.Println("Error adding role:", err)
			return
		}
	}
}

// messageReactionRemove handles removing roles when a reaction is removed from a message.
// @param s the Discord session
// @param m the reaction removal event
func messageReactionRemove(s *discordgo.Session, m *discordgo.MessageReactionRemove) {
	if m.ChannelID != channelRoles {
		return
	}

	if m.MessageID == messageId {

		// Determine the emoji APIName to use
		var roleID string
		if m.Emoji.ID != "" {
			// Custom emoji
			roleID = getRoleID(m.Emoji.ID)
		} else {
			// Standard emoji
			roleID = getRoleID(m.Emoji.Name)
		}

		if roleID == "" {
			return
		}
		err := s.GuildMemberRoleRemove(m.GuildID, m.UserID, roleID)
		if err != nil {
			fmt.Println("Error removing role:", err)
			return
		}
	}
}

// messageTimer processes messages and creates a timer if the message contains a time.
// @param s the Discord session
// @param m the message creation event
func messageTimer(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.ID == "1230113139392122930" {
		return
	}
	// Regular expression to find time durations like "1.5 hours", "1 hour and 15 min", etc.
	re := regexp.MustCompile(`(?i)(\d+[.,]?\d*)\s*(hour|hours|time|timer|minute|minutes|minutter|minutt|min|mins|sec|sek)`)
	matches := re.FindAllStringSubmatch(m.Content, -1)

	// If no matches are found, return
	if len(matches) == 0 {
		return
	}

	// Total duration to accumulate time intervals
	var totalDuration time.Duration

	// Loop through each match and accumulate the total duration
	for _, match := range matches {
		// Replace comma with period to normalize the float number
		numberStr := strings.Replace(match[1], ",", ".", 1)

		// Convert the string to a float
		number, err := strconv.ParseFloat(numberStr, 64)
		if err != nil {
			fmt.Println("Error converting time value:", err)
			return
		}

		unit := strings.ToLower(match[2])

		// Determine the duration based on the unit
		var duration time.Duration
		switch unit {
		case "hour", "hours", "time", "timer":
			duration = time.Duration(number * float64(time.Hour))
		case "minute", "minutter", "minutt", "minutes", "min", "mins":
			duration = time.Duration(number * float64(time.Minute))
		case "sec", "sek":
			duration = time.Duration(number * float64(time.Second))
		}

		// Accumulate the total duration
		totalDuration += duration
	}

	// Format the total duration
	var formattedDuration string
	totalSeconds := int(totalDuration.Seconds())

	if totalSeconds < 60 {
		formattedDuration = fmt.Sprintf("%d seconds", totalSeconds)
	} else if totalSeconds < 3600 {
		formattedDuration = fmt.Sprintf("%.1f minutes", float64(totalSeconds)/60)
	} else {
		formattedDuration = fmt.Sprintf("%.1f hours", totalDuration.Hours())
	}

	// Calculate the Unix timestamp after the total duration
	futureTime := time.Now().Add(totalDuration).Unix()

	// Create the response message in the required format, including the total time
	response := fmt.Sprintf("Created timer for: %s (<t:%d:R>)", formattedDuration, futureTime)

	// Send the response message as a reply
	_, err := s.ChannelMessageSend(m.ChannelID, response)
	if err != nil {
		fmt.Println("Error sending message:", err)
		return
	}

	// Set a timer to send a message when the time is up
	time.AfterFunc(totalDuration, func() {
		// Ping the user who sent the original message
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("<@%s> Time's up bitch!", m.Author.ID))
		if err != nil {
			fmt.Println("Error sending time's up message:", err)
		}
	})
}

// messageCreate processes messages in specific channel and sends the image to openAI.
// @param s the Discord session
// @param m the message creation event
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
			resp.Choices[0].Message.Content = random(gifs)
		}

		_, err = s.ChannelMessageSendReply(m.ChannelID, resp.Choices[0].Message.Content, m.Reference())
		if err != nil {
			fmt.Println("Error sending message:", err)
			return
		}

	}
}

// random returns a random string from a given slice.
// @param array the slice to choose from
// @return string a random element from the slice
func random(array []string) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return array[r.Intn(len(array))]
}

// fetchCommits scrapes and processes commit data from a web page in batches off 10 to a specified discord channel.
// @param s the Discord session
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

			_, err := s.ChannelMessageSendEmbed(rustUpdateChannel, embed)
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

// scrapeSteamStore collects and sends embedded messages about store items for the game Rust.
// @param discord the Discord session
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

// voiceStateUpdate manages voice state updates, potentially creating new voice channels.
// @param dg the Discord session
// @param v the voice state update event
func voiceStateUpdate(dg *discordgo.Session, v *discordgo.VoiceStateUpdate) {
	ctx := context.Background()

	// Start a transaction
	tx, err := dbpool.Begin(ctx)
	if err != nil {
		fmt.Println("Error starting transaction:", err)
		return
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				fmt.Printf("Failed to rollback transaction: %v\n", rollbackErr)
			}
		}
	}()

	var categoryID string
	err = tx.QueryRow(ctx, "SELECT chategoryID FROM mainVoice WHERE channelID = $1", v.ChannelID).Scan(&categoryID)
	if err != nil {
		return
	}

	if categoryID != "" {
		var abbreviation string
		err = tx.QueryRow(ctx, "SELECT abbreviation FROM chategory WHERE chategoryID = $1", categoryID).Scan(&abbreviation)
		if err != nil {
			fmt.Println("Error querying abbreviation from chategory:", err)
			return
		}

		// Fetch all channels from Discord
		channels, err := dg.GuildChannels(v.GuildID) // Assuming v.GuildID is the ID of the guild
		if err != nil {
			fmt.Println("Failed to fetch channels for guild:", err)
			return
		}

		var existingNumbers []int
		prefix := fmt.Sprintf("\U0001F509 | %s Voice ", abbreviation)

		// Iterate over all channels to find existing numbers in channel names
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

		// Fetch parentID from chategory for the new channel creation
		var parentID string
		err = tx.QueryRow(ctx, "SELECT chategoryID FROM chategory WHERE chategoryID = $1", categoryID).Scan(&parentID)
		if err != nil {
			fmt.Println("Category ID not found for the category:", categoryID)
			return
		}

		// Create the new voice channel
		newChannel, err := dg.GuildChannelCreateComplex(v.GuildID, discordgo.GuildChannelCreateData{
			Name:     newChannelName,
			Type:     discordgo.ChannelTypeGuildVoice,
			ParentID: parentID,
		})
		if err != nil {
			fmt.Println("Error creating channel:", err)
			return
		}

		// Insert the new channel into the chiledVoice table
		_, err = tx.Exec(ctx, "INSERT INTO chiledVoice (parentChannelID, channelID) VALUES ($1, $2)", v.ChannelID, newChannel.ID)
		if err != nil {
			fmt.Printf("Error inserting into chiledVoice: %v, parentChannelID: %s, channelID: %s\n", err, v.ChannelID, newChannel.ID)
			return
		}

		// Creates entry for activeVoice for the chiledChannel
		_, err = tx.Exec(ctx, "INSERT INTO activeVoice (channelID) VALUES ($1)", newChannel.ID)
		if err != nil {
			fmt.Printf("Error inserting into activeVoice: %v, channelID: %s\n", err, newChannel.ID)
			return
		}

		// Commit the transaction
		err = tx.Commit(ctx)
		if err != nil {
			fmt.Println("Error committing transaction:", err)
			return
		}

		fmt.Printf("Successfully inserted into chiledVoice: parentChannelID: %s, channelID: %s\n", v.ChannelID, newChannel.ID)

		// Move the user to the new channel
		err = dg.GuildMemberMove(v.GuildID, v.UserID, &newChannel.ID)
		if err != nil {
			fmt.Println("Error moving member:", err)
			return
		}
	}
}

// updateVoiceState processes updates to voice states for maintaining voice channel activity.
// @param dg the Discord session
// @param v the voice state update event
func updateVoiceState(dg *discordgo.Session, v *discordgo.VoiceStateUpdate) {

	ctx := context.Background()

	// Determine the relevant channelID based on whether someone joined or left
	var relevantChannelID string
	if v.BeforeUpdate != nil && v.BeforeUpdate.ChannelID != "" {
		relevantChannelID = v.BeforeUpdate.ChannelID
	} else if v.ChannelID != "" {
		relevantChannelID = v.ChannelID
	} else {
		// No relevant channel to check, just return
		return
	}

	// Check if the relevant channel exists in chiledVoice
	var channelID string
	err := dbpool.QueryRow(ctx, "SELECT channelID FROM chiledVoice WHERE channelID = $1", relevantChannelID).Scan(&channelID)
	if err != nil {
		if err == sql.ErrNoRows {
			return
		}
		return
	}

	// Process voiceUsers increment or decrement based on voice state changes
	if v.ChannelID != "" && (v.BeforeUpdate == nil || v.BeforeUpdate.ChannelID != v.ChannelID) {

		// Increment voiceUsers because user has joined a new channel
		_, err = dbpool.Exec(ctx, "UPDATE activeVoice SET voiceUsers = (voiceUsers + 1) WHERE channelID = $1", v.ChannelID)
		if err != nil {
			fmt.Printf("Error incrementing voiceUsers in channel %s: %v\n", v.ChannelID, err)
		}
	}

	if v.BeforeUpdate != nil && v.BeforeUpdate.ChannelID != "" && v.BeforeUpdate.ChannelID != v.ChannelID {
		// Decrement voiceUsers because user has left a channel

		_, err := dbpool.Exec(ctx, "UPDATE activeVoice SET voiceUsers = (voiceUsers - 1) WHERE channelID = $1", v.BeforeUpdate.ChannelID)
		if err != nil {
			fmt.Printf("Error decrementing voiceUsers in channel %s: %v\n", v.BeforeUpdate.ChannelID, err)
			return
		}

		// Check if voiceUsers is now 0 and delete the channel if so
		var voiceUsers int
		err = dbpool.QueryRow(ctx, "SELECT voiceUsers FROM activeVoice WHERE channelID = $1", v.BeforeUpdate.ChannelID).Scan(&voiceUsers)
		if err != nil {
			fmt.Printf("Error checking voiceUsers in channel %s: %v\n", v.BeforeUpdate.ChannelID, err)
			return
		}

		if voiceUsers <= 0 {
			// Delete the channel from Discord
			if _, err := dg.ChannelDelete(v.BeforeUpdate.ChannelID); err != nil {
				fmt.Printf("Failed to delete channel %s in Discord: %v\n", v.BeforeUpdate.ChannelID, err)
			}
			// Delete the channel from the childVoice table
			_, err = dbpool.Exec(ctx, "DELETE FROM chiledVoice WHERE channelID = $1", v.BeforeUpdate.ChannelID)
			if err != nil {
				fmt.Printf("Error deleting channel from childVoice table %s: %v\n", v.BeforeUpdate.ChannelID, err)
			}
		}
	}
}

// updateRoleMessage updates the message that lists roles in a specific channel.
// @param s the Discord session
// @return error potential error during the update process
func updateRoleMessage(s *discordgo.Session) error {
	const channelID = "1230127864519852104"
	const messageID = "1230141184664535051"

	// Fetch all games from the database
	rows, err := dbpool.Query(context.Background(), "SELECT emoji, name FROM chategory")
	if err != nil {
		return fmt.Errorf("failed to fetch games from database: %w", err)
	}
	defer rows.Close()

	var messageContent strings.Builder
	messageContent.WriteString("Welcome to Trappa!\n\nReact with the following emojis to get the roles:\n")

	for rows.Next() {
		var emoji, name string
		if err := rows.Scan(&emoji, &name); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}
		messageContent.WriteString(fmt.Sprintf("	%s %s\n", emoji, name))
	}

	// Edit the existing role message in the specified channel
	_, err = s.ChannelMessageEdit(channelID, messageID, messageContent.String())
	if err != nil {
		return fmt.Errorf("failed to edit role message: %w", err)
	}

	return nil
}

// acknowledgeInteraction sends an initial response to an interaction.
// @param dg the Discord session
// @param i the interaction event
// @return error potential error during acknowledgment
func acknowledgeInteraction(dg *discordgo.Session, i *discordgo.InteractionCreate) error {
	err := dg.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		log.Printf("Error acknowledging interaction: %v", err)
	}
	return err
}

// sendResponse sends a response to an interaction, handling message length appropriately.
// @param dg the Discord session
// @param i the interaction event
// @param response the response string
// @return error potential error during message sending
func sendResponse(dg *discordgo.Session, i *discordgo.InteractionCreate, response string) error {
	const maxMessageLength = 2000

	if len(response) <= maxMessageLength {
		_, err := dg.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: response,
		})
		if err != nil {
			log.Printf("Error sending follow-up message: %v", err)
		}
		return err
	}

	for k := 0; k < len(response); k += maxMessageLength {
		end := k + maxMessageLength
		if end > len(response) {
			end = len(response)
		}

		chunk := response[k:end]

		_, err := dg.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: chunk,
		})
		if err != nil {
			log.Printf("Error sending follow-up message: %v", err)
			return err
		}
	}

	return nil
}

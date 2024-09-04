# Trappabot
Trappa bot is a Discord bot made to manage a discord server me and my friends use. The bot will give and remove roles, create and delete voice channels and roles, scrape webpages for information, talk to OpenAI and allot more!

Warning: The DLL i used to interact with discords API, discordgo (https://github.com/bwmarrin/discordgo). I could not make it work when i had the discord interaction between several files. So EVERYTHING is in the main.go. However, the code is commented with doxygen.

## Useing the bot

### Before use
The bot uses a database that it need access to before the bot will be able to start. The SQL of the database is at the end of the file. You will also need these values, ive added them to a .env file personaly. The code will also be able to find a .env file and use them en enviorment varables.

```
DISCORD_BOT_TOKEN=""

OPENAI_API_KEY=""

ADMIN_ID=""

DB_USER=""
DB_PASSWORD=""
DB_HOST=""
DB_PORT=""
DB_NAME=""
```

you will also need to change these values at the top of the file:

var messageId = "1230141184664535051"

var channelRoles = "1230127864519852104"

var channelReview = "1013473566806786058"

var GuildID = "1012016741238448278"

var rustUpdateChannel = "1243953919621861428"


channelRoles and messageId is for the message in the channel people can interact with to get given roles based on games.

channelReview is for the channel the bot will react to how deadly the staircase you'll send in an image is. 

GuildID is the ID of the discord server.

rustUpdateChannel is for the channel where the bot will check every 2 min for changes to the game rust's change log.

### Commands
The / commands the bot uses. (this message is also the reponse to /help)

Ｃｏｍｍａｎｄｓ:

**/event 'ping' 'event description' 'event start time' optional: 'person limit'**: This command will create a message that people can react to. This will notify the participants when the event start. You can also set the max amout of people allowed at the event!
		
**/help**: Responds with _this_ message.
		 
Ａｄｍｉｎ Ｃｏｍｍａｎｄｓ:

**/creategame 'name of the game' 'name abbreviation' 'server emoji for the game'**: This command will create everything needed to add a new game to the discord. Channels, roles, permissions and so on!
		
**/deletegame 'name of the game'**: Will delete the game and everything associating with the game. Channels, roles, permissions and so on!

**/purge**: Deletes all chiled channels!


## Database

Table chategory
roleID, name, emoji, chategoryID, abbriviation

Table mainVoice
chategoryID, channelID

table chiledVoice
parentChannelID, channelID

Table activeVoice
channelID, voiceUsers


chategory.chategoryID = mainVoice.chategoryID
mainVoice.channelID = chiledVoice.parentChannelID
chiledVoice.channelID = activeVoice.ChannelID

### SQL statements

-- Create chategory table
CREATE TABLE chategory (
    chategoryID VARCHAR(255) PRIMARY KEY,
    roleID VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    emoji VARCHAR(255) NOT NULL,
    abbreviation VARCHAR(255) NOT NULL,
    UNIQUE (name),
    UNIQUE (emoji)
);

-- Create mainVoice table
CREATE TABLE mainVoice (
    chategoryID VARCHAR(255),
    channelID VARCHAR(255) NOT NULL,
    PRIMARY KEY (chategoryID, channelID),
    UNIQUE (channelID)
);

-- Create chiledVoice table
CREATE TABLE chiledVoice (
    parentChannelID VARCHAR(255),
    channelID VARCHAR(255) NOT NULL,
    PRIMARY KEY (parentChannelID, channelID),
    UNIQUE (channelID)
);

-- Create activeVoice table
CREATE TABLE activeVoice (
    channelID VARCHAR(255) PRIMARY KEY,
    voiceUsers INTEGER NOT NULL DEFAULT 1
);

-- Add foreign key constraints
ALTER TABLE mainVoice
    ADD CONSTRAINT fk_mainVoice_chategory
    FOREIGN KEY (chategoryID)
    REFERENCES chategory (chategoryID)
    ON DELETE CASCADE;

ALTER TABLE chiledVoice
    ADD CONSTRAINT fk_chiledVoice_mainVoice
    FOREIGN KEY (parentChannelID)
    REFERENCES mainVoice (channelID)
    ON DELETE CASCADE;

ALTER TABLE activeVoice
    ADD CONSTRAINT fk_activeVoice_chiledVoice
    FOREIGN KEY (channelID)
    REFERENCES chiledVoice (channelID)
    ON DELETE CASCADE;

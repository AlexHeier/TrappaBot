# Trappabot
Trappa bot is a Discord bot made to manage a discord server me and my friends use. The bot will give and remove roles, create and delete voice channels and roles, scrape webpages for information, talk to OpenAI and allot more!

Warning: The DLL i used to interact with discords API, discordgo (https://github.com/bwmarrin/discordgo). I could not make it work when i had the discord interaction doxygen between several files. So EVERYTHING is in the main.go. However, the code is commented with doxygen.

## Commands

Ｃｏｍｍａｎｄｓ:
**/event <ping> <event description> <event start time> optional: <person limit>**: This command will create a message that people can react to. This will notify the participants when the event start. You can also set the max amout of people allowed at the event!
		
**/help**: Responds with _this_ message.
		 
Ａｄｍｉｎ Ｃｏｍｍａｎｄｓ:
**/creategame <name of the game> <name abbreviation> <server emoji for the game>**: This command will create everything needed to add a new game to the discord. Channels, roles, permissions and so on!
		
**/deletegame <name of the game>**: Will delete the game and everything associating with the game. Channels, roles, permissions and so on!

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

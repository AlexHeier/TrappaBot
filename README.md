# Trappabot

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

-------------------------------

DROP TABLE IF EXISTS activeVoice;
DROP TABLE IF EXISTS chiledVoice;
DROP TABLE IF EXISTS mainVoice;
DROP TABLE IF EXISTS chategory;

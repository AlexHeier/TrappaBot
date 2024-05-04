package functions

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
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
}

var messageId = "1230141184664535051"
var channelRoles = "1230127864519852104"

func getRoleID(emoji string) string {
	roleID, ok := emojiRoleMap[emoji]
	if !ok {
		return ""
	}
	return roleID
}

func MessageReactionAdd(s *discordgo.Session, m *discordgo.MessageReactionAdd) {

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

func MessageReactionRemove(s *discordgo.Session, m *discordgo.MessageReactionRemove) {

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

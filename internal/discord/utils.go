package discord

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

// GetMemberColor calculates the color of a member based on their roles.
// Returns a hex string (e.g. "#ff0000") or empty string if no color.
func GetMemberColor(guild *discordgo.Guild, member *discordgo.Member) string {
	if guild == nil || member == nil {
		return ""
	}

	roles := make(map[string]*discordgo.Role)
	for _, r := range guild.Roles {
		roles[r.ID] = r
	}

	var colorRole *discordgo.Role

	for _, roleID := range member.Roles {
		if role, ok := roles[roleID]; ok {
			if role.Color != 0 {
				if colorRole == nil || role.Position > colorRole.Position {
					colorRole = role
				}
			}
		}
	}

	if colorRole != nil {
		return fmt.Sprintf("#%06x", colorRole.Color)
	}
	return ""
}

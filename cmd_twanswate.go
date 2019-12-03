package commandworks

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	pbs "github.com/cytobot/messaging/transport/shared"
	"github.com/lampjaw/discordclient"
)

type twanswateCommandProcessor struct {
	discord *discordResolver
}

func newTwanswateCommandProcessor(discordResolver *discordResolver) *twanswateCommandProcessor {
	return &twanswateCommandProcessor{
		discord: discordResolver,
	}
}

func (p *twanswateCommandProcessor) getCommandID() string {
	return "twanswate"
}

func (p *twanswateCommandProcessor) process(client *discordclient.DiscordClient, req *pbs.DiscordWorkRequest) error {
	previousMessages, err := client.GetMessages(req.ChannelID, 1, req.MessageID)
	if err != nil {
		client.SendMessage(req.ChannelID, fmt.Sprintf("%s", err))
		return err
	}

	if previousMessages == nil || len(previousMessages) == 0 {
		client.SendMessage(req.ChannelID, "Unable to find a message to translate.")
		return nil
	}

	var previousMessage = previousMessages[0]

	if client.IsMe(previousMessage) {
		return nil
	}

	replacer := strings.NewReplacer(
		"r", "w",
		"R", "W",
		"l", "w",
		"L", "W")

	translatedText := replacer.Replace(previousMessage.Message())

	if err != nil {
		client.SendMessage(req.ChannelID, fmt.Sprintf("%s", err))
		return nil
	}

	channel, err := p.discord.ResolveChannel(req.SourceID, req.ChannelID)
	guild, err := p.discord.ResolveGuild(req.SourceID, req.GuildID)
	user, err := p.discord.ResolveUser(req.SourceID, previousMessage.UserID(), previousMessage.Channel())

	timestamp, err := previousMessage.Timestamp()

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    user.Username,
			IconURL: discordgo.EndpointUserAvatar(user.ID, user.Avatar),
		},
		Color:       0x070707,
		Description: translatedText,
		Timestamp:   timestamp.UTC().Format("2006-01-02T15:04:05-0700"),
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("in #%s at %s", channel.Name, guild.Name),
		},
	}

	client.SendEmbedMessage(req.ChannelID, embed)

	return nil
}

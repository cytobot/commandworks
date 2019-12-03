package commandworks

import (
	"fmt"

	pbs "github.com/cytobot/messaging/transport/shared"
	"github.com/lampjaw/discordclient"
)

type inviteCommandProcessor struct{}

func newInviteCommandProcessor() *inviteCommandProcessor {
	return &inviteCommandProcessor{}
}

func (p *inviteCommandProcessor) getCommandID() string {
	return "invite"
}

func (p *inviteCommandProcessor) process(client *discordclient.DiscordClient, req *pbs.DiscordWorkRequest) error {
	if client.ClientID != "" {
		msg := fmt.Sprintf("Please visit <https://discordapp.com/oauth2/authorize?client_id=%s&scope=bot> to add %s to your server.", client.ClientID, client.UserName())
		client.SendMessage(req.ChannelID, msg)
	}

	return nil
}

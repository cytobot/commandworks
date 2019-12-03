package commandworks

import (
	"fmt"

	pbs "github.com/cytobot/messaging/transport/shared"
	"github.com/lampjaw/discordclient"
)

type command interface {
	// GetCommandID returns the unique command identifier
	getCommandID() string
	// Process executes the command
	process(client *discordclient.DiscordClient, req *pbs.DiscordWorkRequest) error
}

type CommandHandler struct {
	commands []command
}

func NewCommandHandler() *CommandHandler {
	return &CommandHandler{
		commands: getCommandTypes(),
	}
}

func (h *CommandHandler) ProcessCommand(client *discordclient.DiscordClient, req *pbs.DiscordWorkRequest) error {
	cmd := h.findCommand(req.Command)
	if cmd == nil {
		return fmt.Errorf("[CommandHandler] Command with identifier '%s' was not found", req.Command)
	}

	err := cmd.process(client, req)
	if err != nil {
		return fmt.Errorf("[CommandHandler] Command '%s' failed: %v", req.Command, err)
	}

	return nil
}

func (h *CommandHandler) findCommand(commandID string) command {
	for _, cmd := range h.commands {
		if cmd.getCommandID() == commandID {
			return cmd
		}
	}
	return nil
}

func getCommandTypes() []command {
	return []command{
		newInviteCommandProcessor(),
		newTwanswateCommandProcessor(),
	}
}

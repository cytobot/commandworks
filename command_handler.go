package commandworks

import (
	"fmt"
	"log"

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
	commands      []command
	managerclient *managerClient
}

func NewCommandHandler(managerEndpoint string) *CommandHandler {
	handler := &CommandHandler{
		managerclient: getManagerClient(managerEndpoint),
	}
	handler.commands = handler.getCommandTypes()
	return handler
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

func getManagerClient(managerEndpoint string) *managerClient {
	client, err := newManagerClient(managerEndpoint)
	if err != nil {
		panic(fmt.Sprintf("[Manager client error] %s", err))
	}

	log.Println("Connected to manager client")

	return client
}

func (h *CommandHandler) findCommand(commandID string) command {
	for _, cmd := range h.commands {
		if cmd.getCommandID() == commandID {
			return cmd
		}
	}
	return nil
}

func (h *CommandHandler) getCommandTypes() []command {
	return []command{
		newInviteCommandProcessor(),
		newTwanswateCommandProcessor(),
		newStatsCommandProcessor(h.managerclient),
	}
}

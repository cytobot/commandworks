package processor

import (
	"fmt"

	"github.com/cytobot/commandworks/discord"
)

// CommandDefinition is the basic type for defining plugin commands
type CommandDefinition struct {
	// CommandID is an internal id used for internal tracking.
	CommandID string
	// ProcessFunc is a function reference that's called when a message meets trigger and argument requirements.
	ProcessFunc func(bot *Processor, client *discord.DiscordClient, payload CommandPayload)
}

// CommandPayload contains data related to the incomming request
type CommandPayload struct {
	//CommandID
	CommandID string
	// Trigger is the specific string that activated the command
	Trigger string
	// Arguments contain a hash of all configured CommandDefinitionArguments that could be parsed
	Arguments map[string]string
	// Message is the entire message received that activated the command
	Message discord.Message
}

// IsValid determines if the command definition is configured correctly
func (c *CommandDefinition) IsValid() (bool, []string) {
	errors := make([]string, 0)

	if c.CommandID == "" {
		errors = append(errors, "No CommandID provided for CommandDefinition")
	}

	if c.ProcessFunc == nil {
		errors = append(errors, "No callback provided for CommandDefinition")
	}

	return len(errors) > 0, errors
}

func validateCommand(command *CommandDefinition) bool {
	errors := make([]string, 0)

	if isValid, commandErrors := command.IsValid(); !isValid {
		errors = append(errors, commandErrors...)
	}

	if len(errors) > 0 {
		for _, errmsg := range errors {
			fmt.Printf("Command validation error: %s: %s", command.CommandID, errmsg)
		}
		return false
	}

	return true
}

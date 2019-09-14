package listener

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"

	"github.com/cytobot/commandworks/discord"
)

// ListenerConf defines operational parameters to be used with NewBot
type ListenerConf struct {
	// CommandPrefix is a string thats prefixed to every command trigger.
	CommandPrefix string
	// ClientID sets the known client id of the bot. Potentially useful in some plugins.
	ClientID string
	// OwnerUserID is the OwnerUserId. Needed for processing commands restricted to the owner permission.
	OwnerUserID string
	// CommandCallback is the callback for every command
	CommandCallback func(bot *Listener, client *discord.DiscordClient, payload CommandPayload)
}

// Listener handles bot related functionality
type Listener struct {
	Client          *discord.DiscordClient
	Plugins         map[string]IPlugin
	Commands        map[string]*CommandDefinition
	Config          *ListenerConf
	messageChannels []chan discord.Message
	State           interface{}
}

// Open loads plugin data and starts listening for discord messages
func (b *Listener) Open() {
	var invalidPlugin = false
	var invalidCommand = false

	for _, plugin := range b.Plugins {
		if !validatePlugin(plugin) {
			invalidPlugin = true
		}
	}

	if invalidPlugin {
		log.Printf("A misconfigured plugin was found.")
		return
	}

	for _, command := range b.Commands {
		if !validateCommand(command) {
			invalidCommand = true
		}
	}

	if invalidCommand {
		log.Printf("A misconfigured command was found.")
		return
	}

	if messageChan, err := b.Client.Open(); err == nil {
		go b.listen(messageChan)
	} else {
		log.Printf("Error creating discord service: %v\n", err)
	}
}

// RegisterPlugin registers a plugin to process messages or commands
func (b *Listener) RegisterPlugin(plugin IPlugin) {
	if b.Plugins[plugin.Name()] != nil {
		log.Println("Plugin with that name already registered", plugin.Name())
	}
	b.Plugins[plugin.Name()] = plugin
}

// GetCommandPrefix returns the prefix as configured in the ListenerConf or the default if none is available
func (b *Listener) GetCommandPrefix() string {
	return b.Config.CommandPrefix
}

func (b *Listener) listen(messageChan <-chan discord.Message) {
	log.Printf("Listening")
	for {
		message := <-messageChan

		commandPrefix := b.GetCommandPrefix()

		if !strings.HasPrefix(message.Message(), commandPrefix) || message.Message() == "" || b.Client.IsMe(message) {
			continue
		}

		if isCommandsRequest(b.Client, commandPrefix, message) {
			go handleCommandsRequest(b, message, commandPrefix)
			continue
		}

		messageParts := strings.Fields(message.RawMessage())

	pluginLoop:
		for _, plugin := range b.Plugins {
			if plugin.Commands() == nil {
				continue
			}

			for _, commandDefinition := range plugin.Commands() {
				if isMatch := findCommandDefinitionCommandMatch(b, commandDefinition, message, commandPrefix, messageParts); isMatch {
					break pluginLoop
				}
			}
		}
	}
}

func findCommandDefinitionCommandMatch(b *Listener, commandDefinition *CommandDefinition, message discord.Message, commandPrefix string, parts []string) bool {
	if !validateCommandAccess(b.Client, commandDefinition, message) {
		return false
	}

	for _, trigger := range commandDefinition.Triggers {
		if isTriggerMatch, triggerMatch := findTriggerMatch(commandDefinition, trigger, commandPrefix, parts, message); isTriggerMatch {
			if isArgumentMatch, parsedArgs := extractCommandArguments(message, triggerMatch, commandDefinition.Arguments); isArgumentMatch {
				log.Printf("<%s> %s: %s\n", message.Channel(), message.UserName(), message.RawMessage())

				payload := CommandPayload{
					CommandID: commandDefinition.CommandID,
					Trigger:   trigger,
					Arguments: parsedArgs,
					Message:   message,
				}

				go b.Config.CommandCallback(b, b.Client, payload)
				return true
			}
		}
	}

	return false
}

func findTriggerMatch(commandDefinition *CommandDefinition, commandTrigger string, definitionPrefix string, messageParts []string, message discord.Message) (bool, string) {
	if messageParts[0] == definitionPrefix+commandTrigger {
		return true, messageParts[0]
	}

	if !commandDefinition.DisableTriggerOnMention && len(messageParts) > 1 {
		return message.IsMentionTrigger(commandTrigger)
	}

	return false, ""
}

func validateCommandAccess(client *discord.DiscordClient, commandDefinition *CommandDefinition, message discord.Message) bool {
	if commandDefinition.ExposureLevel > 0 {
		switch commandDefinition.ExposureLevel {
		case EXPOSURE_PRIVATE:
			if !client.IsPrivate(message) {
				return false
			}
		case EXPOSURE_PUBLIC:
			if client.IsPrivate(message) {
				return false
			}
		}
	}

	return validateCommandAccessPermission(client, commandDefinition.PermissionLevel, message)
}

func validateCommandAccessPermission(client *discord.DiscordClient, permissionLevel PermissionLevel, message discord.Message) bool {
	if permissionLevel <= 0 {
		return true
	}

	switch permissionLevel {
	case PERMISSION_USER:
		return true
	case PERMISSION_MODERATOR:
		if client.IsModerator(message) {
			return true
		}
		fallthrough
	case PERMISSION_ADMIN:
		if client.IsChannelOwner(message) {
			return true
		}
		fallthrough
	case PERMISSION_OWNER:
		if client.IsBotOwner(message) {
			return true
		}
	}

	return false
}

func extractCommandArguments(message discord.Message, trigger string, arguments []CommandDefinitionArgument) (bool, map[string]string) {
	parsedArgs := make(map[string]string)

	if arguments == nil || len(arguments) == 0 {
		return true, parsedArgs
	}

	var argPatterns []string

	for i, argument := range arguments {
		pattern := ""

		if i == 0 {
			pattern = fmt.Sprintf("(?P<%s>%s)", argument.Alias, argument.Pattern)
		} else {
			pattern = fmt.Sprintf("(?:\\s+(?P<%s>%s))", argument.Alias, argument.Pattern)
		}

		if argument.Optional {
			pattern += "?"
		}

		argPatterns = append(argPatterns, pattern)
	}
	var pattern = fmt.Sprintf("^%s$", strings.Join(argPatterns, ""))

	var trimmedContent = strings.TrimSpace(strings.TrimPrefix(message.RawMessage(), fmt.Sprintf("%s", trigger)))
	pat := regexp.MustCompile(pattern)
	argsMatch := pat.FindStringSubmatch(trimmedContent)

	if len(argsMatch) == len(arguments)-1 && arguments[len(arguments)-1].Optional {
		argsMatch = append(argsMatch, "")
	}

	if argsMatch == nil || len(argsMatch) == 1 {
		return false, nil
	}

	for i := 1; i < len(argsMatch); i++ {
		parsedArgs[pat.SubexpNames()[i]] = argsMatch[i]
	}

	if len(parsedArgs) != len(arguments) {
		return false, nil
	}

	return true, parsedArgs
}

func handleCommandsRequest(b *Listener, message discord.Message, commandPrefix string) {
	help := []string{}

	for _, plugin := range b.Plugins {
		var h []string

		helpResult := plugin.Help(b, b.Client, message, false)

		if helpResult != nil {
			h = helpResult
		} else if plugin.Commands() != nil {
			for _, commandDefinition := range plugin.Commands() {
				if commandDefinition.Unlisted {
					continue
				}

				h = append(h, commandDefinition.Help(b.Client, commandPrefix))
			}
		}

		if h != nil && len(h) > 0 {
			help = append(help, h...)
		}
	}

	for _, commandDefinition := range b.Commands {
		if commandDefinition.Unlisted {
			continue
		}

		help = append(help, commandDefinition.Help(b.Client, commandPrefix))
	}

	sort.Strings(help)

	if len(help) == 0 {
		help = []string{"No commands found"}
	}

	b.Client.SendMessage(message.Channel(), strings.Join(help, "\n"))
}

func isCommandsRequest(client *discord.DiscordClient, commandPrefix string, message discord.Message) bool {
	triggerTerm := "commands"
	commandTrigger := commandPrefix + triggerTerm

	triggered, _ := message.IsMentionTrigger(triggerTerm)

	return triggered || strings.HasPrefix(message.RawMessage(), commandTrigger)
}

package processor

import (
	"log"

	"github.com/cytobot/commandworks/discord"
)

// Listener handles bot related functionality
type Processor struct {
	Client   *discord.DiscordClient
	Plugins  map[string]IPlugin
	Commands map[string]*CommandDefinition
	State    interface{}
}

// Open loads plugin data and starts listening for discord messages
func (b *Processor) Open(messageChan <-chan CommandPayload) {
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

	go b.listen(messageChan)
}

// RegisterPlugin registers a plugin to process messages or commands
func (b *Processor) RegisterPlugin(plugin IPlugin) {
	if b.Plugins[plugin.Name()] != nil {
		log.Println("Plugin with that name already registered", plugin.Name())
	}
	b.Plugins[plugin.Name()] = plugin
}

func (b *Processor) listen(messageChan <-chan CommandPayload) {
	log.Printf("Awaiting work")
	for {
		payload := <-messageChan

	pluginLoop:
		for _, plugin := range b.Plugins {
			if plugin.Commands() == nil {
				continue
			}

			for _, commandDefinition := range plugin.Commands() {
				if commandDefinition.CommandID == payload.CommandID {
					go commandDefinition.ProcessFunc(b, b.Client, payload)
					break pluginLoop
				}
			}
		}
	}
}

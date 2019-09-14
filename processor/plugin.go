package processor

import (
	"fmt"
	"sync"
)

// IPlugin is the universal required interface for all plugins
type IPlugin interface {
	// Name returns the name of the plugin
	Name() string
	// Commands returns an array of CommandDefinitions
	Commands() []*CommandDefinition
}

// Plugin is the basic model to build bot plugins off of
type Plugin struct {
	sync.RWMutex
}

// Commands returns an array of CommandDefinitions
func (p *Plugin) Commands() []*CommandDefinition {
	return nil
}

// Name returns the name of the plugin
func (p *Plugin) Name() string {
	return ""
}

func validatePlugin(plugin IPlugin) bool {
	errors := make([]string, 0)

	if plugin.Name() == "" {
		errors = append(errors, "Plugin validation error: Missing required Name")
	}

	for _, command := range plugin.Commands() {
		if isValid, commandErrors := command.IsValid(); !isValid {
			errors = append(errors, commandErrors...)
		}
	}

	if len(errors) > 0 {
		for _, errmsg := range errors {
			fmt.Printf("Plugin validation error: %s: %s", plugin.Name(), errmsg)
		}
		return false
	}

	return true
}
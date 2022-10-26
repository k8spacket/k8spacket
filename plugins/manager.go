package plugins

import (
	"github.com/k8spacket/plugin-api"
	"net/http"
)

type Manager struct {
	streamPlugins []plugin_api.StreamPlugin
}

func NewPluginManager() *Manager {
	pm := &Manager{}
	return pm
}

func (pm *Manager) RegisterPlugin(streamPlugin plugin_api.StreamPlugin) {
	pm.streamPlugins = append(pm.streamPlugins, streamPlugin.(plugin_api.StreamPlugin))
}

func (pm *Manager) RegisterHttpHandler(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc(pattern, handler)
}

func (pm *Manager) GetStreamPlugins() []plugin_api.StreamPlugin {
	return pm.streamPlugins
}

package plugins

import (
	"github.com/k8spacket/plugin-api/v2"
	"net/http"
)

type Manager struct {
	tcpConsumerPlugins []plugin_api.TCPConsumerPlugin
	tlsConsumerPlugins []plugin_api.TLSConsumerPlugin
}

func NewPluginManager() *Manager {
	pm := &Manager{}
	return pm
}

func (pm *Manager) RegisterTCPPlugin(consumerPlugin plugin_api.TCPConsumerPlugin) {
	pm.tcpConsumerPlugins = append(pm.tcpConsumerPlugins, consumerPlugin.(plugin_api.TCPConsumerPlugin))
}

func (pm *Manager) RegisterTLSPlugin(consumerPlugin plugin_api.TLSConsumerPlugin) {
	pm.tlsConsumerPlugins = append(pm.tlsConsumerPlugins, consumerPlugin.(plugin_api.TLSConsumerPlugin))
}

func (pm *Manager) RegisterHttpHandler(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc(pattern, handler)
}

func (pm *Manager) GetTCPConsumerPlugins() []plugin_api.TCPConsumerPlugin {
	return pm.tcpConsumerPlugins
}

func (pm *Manager) GetTLSConsumerPlugins() []plugin_api.TLSConsumerPlugin {
	return pm.tlsConsumerPlugins
}

package plugins

import (
	k8spacket_log "github.com/k8spacket/k8spacket/log"
	"github.com/k8spacket/plugin-api/v2"
	"io/fs"
	"os"
	"path/filepath"
	"plugin"
)

func InitPlugins(manager *Manager) {

	for _, pluginPath := range find("./plugins", ".so") {

		plug, err := plugin.Open(pluginPath)
		if err != nil {
			k8spacket_log.LOGGER.Printf("[plugins] Cannot open plugins path, %+v", err)
			os.Exit(1)
		}

		initPlugin(plug, manager, "TCPConsumerPlugin")
		initPlugin(plug, manager, "TLSConsumerPlugin")

	}
}

func find(root, ext string) []string {
	var a []string
	filepath.WalkDir(root, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if filepath.Ext(d.Name()) == ext {
			a = append(a, s)
		}
		return nil
	})
	return a
}

func initPlugin(plug *plugin.Plugin, manager *Manager, symName string) {
	found, err := plug.Lookup(symName)
	if err != nil {
		k8spacket_log.LOGGER.Printf("[plugins] Cannot find plugin %s, gave up. %+v", symName, err)
		return
	}
	var consumerPlugin plugin_api.ConsumerPlugin
	consumerPlugin, ok := found.(plugin_api.ConsumerPlugin)
	if !ok {
		k8spacket_log.LOGGER.Println("[plugins] Unexpected type from module symbol")
		os.Exit(1)
	}
	consumerPlugin.InitPlugin(manager)
}

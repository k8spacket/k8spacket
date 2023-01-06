package plugins

import (
	"fmt"
	"github.com/k8spacket/plugin-api"
	"io/fs"
	"os"
	"path/filepath"
	"plugin"
)

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

func InitPlugins(manager *Manager) {

	for _, pluginPath := range find("./plugins", ".so") {

		println(pluginPath)

		plug, err := plugin.Open(pluginPath)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		symStreamer, err := plug.Lookup("StreamPlugin")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		var streamPlugin plugin_api.StreamPlugin
		streamPlugin, ok := symStreamer.(plugin_api.StreamPlugin)
		if !ok {
			fmt.Println("unexpected type from module symbol")
			os.Exit(1)
		}

		streamPlugin.InitPlugin(manager)
	}
}

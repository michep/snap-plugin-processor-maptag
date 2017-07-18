package main

import (
	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	"github.com/michep/snap-plugin-processor-maptag/maptag"
)

const (
	pluginName    = "maptag"
	pluginVersion = 1
)

func main() {
	plugin.StartProcessor(maptag.NewPlugin(), pluginName, pluginVersion)
}

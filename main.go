package main

import (
	log "github.com/hashicorp/go-hclog"

	"github.com/nskerl/nomad-driver-winsvc/winsvc"

	"github.com/hashicorp/nomad/plugins"
)

func main() {
	// Serve the plugin
	plugins.Serve(factory)
}

// factory returns a new instance of a nomad driver plugin
func factory(log log.Logger) interface{} {
	return winsvc.NewWinsvcDriver(log)
}

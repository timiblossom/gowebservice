package main

import (
	"log"
	"os"
	"runtime"

	"app/provider"
	"app/route"
	"app/shared/config"
	"app/shared/database"
	"app/shared/jsonconfig"
	"app/shared/server"
	"app/shared/session"
	"app/shared/view"
	"app/shared/view/plugin"
)

// *****************************************************************************
// Application Logic
// *****************************************************************************
func init() {
	// Verbose logging with file name and line number
	log.SetFlags(log.Lshortfile)

	// Use all CPU cores
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	// Load the configuration file
	jsonconfig.Load("config"+string(os.PathSeparator)+"config.json", config.Config)

	// Configure the session cookie store
	session.Configure(config.Session())

	// Connect to database
	database.Connect(config.Database())

	// After database connect - init application cache
	provider.InitAppCache()

	// Setup the views
	view.Configure(config.View())
	view.LoadTemplates(config.Template().Root, config.Template().Children)
	view.LoadPlugins(
		plugin.TagHelper(config.View()),
		plugin.NoEscape(),
		plugin.PrettyTime())

	// Start the listener
	server.Run(route.LoadHTTP(), route.LoadHTTPS(), config.Server())
}

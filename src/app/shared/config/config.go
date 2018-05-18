package config

import (
	"encoding/json"

	"app/shared/awss3"
	"app/shared/database"
	"app/shared/email"
	"app/shared/server"
	"app/shared/session"
	"app/shared/view"
)

// *****************************************************************************
// Application Settings
// *****************************************************************************

// Config the settings variable
var Config = &Configuration{}

// Configuration contains the application settings
type Configuration struct {
	Database database.Info     `json:"Database"`
	Email    email.SMTPInfo    `json:"Email"`
	Server   server.Server     `json:"Server"`
	Session  session.Session   `json:"Session"`
	Template view.Template     `json:"Template"`
	View     view.View         `json:"View"`
	AWSS3    awss3.AWSS3Config `json:"AWSS3Config"`
}

// ParseJSON unmarshals bytes to structs
func (c *Configuration) ParseJSON(b []byte) error {
	return json.Unmarshal(b, &c)
}

// Session return current sessions store from config
func Session() session.Session {
	return Config.Session
}

// Database return current database connection
func Database() database.Info {
	return Config.Database
}

// View return current view settings
func View() view.View {
	return Config.View
}

// Server return current server settings
func Server() server.Server {
	return Config.Server
}

// Template return current template settings
func Template() view.Template {
	return Config.Template
}

// AWSS3 return Amazon S3 config settings
func AWSS3() awss3.AWSS3Config {
	return Config.AWSS3
}

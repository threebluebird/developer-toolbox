package app

import (
	"context"
	"fmt"

	"developer-toolbox/backend/config"
	"developer-toolbox/backend/models"
)

// App is the root application container for the toolbox.
type App struct {
	ctx  context.Context
	name string
}

// NewApp creates a new application instance.
func NewApp() *App {
	return &App{name: config.AppName}
}

// Name returns the application name.
func (a *App) Name() string {
	return a.name
}

// Greet returns a friendly greeting for the given user.
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, welcome to %s!", name, a.name)
}

// GetDefaultSettings returns the default runtime settings.
func (a *App) GetDefaultSettings() models.Settings {
	return config.DefaultSettings()
}

// Health returns a simple health payload so the UI can validate the app is alive.
func (a *App) Health() map[string]any {
	return map[string]any{
		"status":  "ok",
		"name":    a.name,
		"version": "0.1.0",
	}
}

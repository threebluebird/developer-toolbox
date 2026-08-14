package app

import "context"

// Startup is called by Wails on application start.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

// Shutdown is called when the application is closing.
func (a *App) Shutdown() {
	// Reserved for future lifecycle hooks.
}

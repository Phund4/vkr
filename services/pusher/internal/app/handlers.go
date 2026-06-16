package app

import (
	httpHandler "pusher/internal/adapters/http/handlers"
)

const (
	HealthPath      = "/health"
	ProbeReadyPath  = "/probe/ready"
	ProbeHealthPath = "/probe/health"
)

type handlers struct {
	probe *httpHandler.ProbeHandler
}

func newHandlers() *handlers {
	return &handlers{
		probe: httpHandler.NewProbeHandler(),
	}
}

func (app *App) initHandlers() {
	app.httpServer.GET(HealthPath, app.deps.handlers.probe.Health)
	app.httpServer.GET(ProbeHealthPath, app.deps.handlers.probe.ProbeHealth)
	app.httpServer.GET(ProbeReadyPath, app.deps.handlers.probe.Ready)
}

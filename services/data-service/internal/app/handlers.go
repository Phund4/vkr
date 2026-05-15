package app

import (
	"net/http"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"

	httpHandler "data-service/internal/adapters/http/handlers"
	"data-service/internal/core/services"
)

const (
	HealthPath          = "/health"
	ProbeReadyPath      = "/probe/ready"
	ProbeHealthPath     = "/probe/health"
	SwaggerPathPrefix   = "/swagger"
	SwaggerPathWildcard = "/swagger/*"
	SwaggerIndexHTML    = "/swagger/index.html"

	ApiV1Path = "/api/v1"

	RoadIncidentsPath  = "/road_incidents"
	RoadCongestionPath = "/road_congestion"
)

type handlers struct {
	probe *httpHandler.ProbeHandler
	road  *httpHandler.RoadDataHandler
}

// GetHandlers создаёт набор HTTP-обработчиков приложения.
func GetHandlers(svc *services.RoadDataService) *handlers {
	return &handlers{
		probe: httpHandler.NewProbeHandler(),
		road:  httpHandler.NewRoadDataHandler(svc),
	}
}

// initHandlers регистрирует HTTP-роуты и swagger endpoints.
func (app *App) initHandlers() {
	app.httpServer.GET(SwaggerPathPrefix, func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, SwaggerIndexHTML)
	})
	app.httpServer.GET(SwaggerPathWildcard, echoSwagger.WrapHandler)
	app.httpServer.GET(HealthPath, app.deps.handlers.probe.Health)
	app.httpServer.GET(ProbeHealthPath, app.deps.handlers.probe.ProbeHealth)
	app.httpServer.GET(ProbeReadyPath, app.deps.handlers.probe.Ready)

	apiV1 := app.httpServer.Group(ApiV1Path)
	apiV1.GET(RoadIncidentsPath, app.deps.handlers.road.ListRoadIncidents)
	apiV1.GET(RoadCongestionPath, app.deps.handlers.road.ListRoadCongestion)
}

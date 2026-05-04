package app

import (
	"net/http"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"

	httpHandler "data-service/internal/adapters/http/handlers"
)

const (
	HealthPath          = "/health"
	ProbeReadyPath      = "/probe/ready"
	ProbeHealthPath     = "/probe/health"
	SwaggerPathPrefix   = "/swagger"
	SwaggerPathWildcard = "/swagger/*"
	SwaggerIndexHTML    = "/swagger/index.html"

	ApiV1Path = "/api/v1"

	GetProductPath = "/get_product"
)

type handlers struct {
	probe   *httpHandler.ProbeHandler
	product *httpHandler.ProductHandler
}

// GetHandlers создает набор HTTP-обработчиков приложения.
func GetHandlers(service httpHandler.ProductGetter) *handlers {
	return &handlers{
		probe:   httpHandler.NewProbeHandler(),
		product: httpHandler.NewProductHandler(service),
	}
}

// initHandlers регистрирует HTTP-роуты и swagger endpoints.
func (app *App) initHandlers() {
	// Технические ручки
	app.httpServer.GET(SwaggerPathPrefix, func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, SwaggerIndexHTML)
	})
	app.httpServer.GET(SwaggerPathWildcard, echoSwagger.WrapHandler)
	app.httpServer.GET(HealthPath, app.deps.handlers.probe.Health)
	app.httpServer.GET(ProbeHealthPath, app.deps.handlers.probe.ProbeHealth)
	app.httpServer.GET(ProbeReadyPath, app.deps.handlers.probe.Ready)

	apiV1 := app.httpServer.Group(ApiV1Path)
	apiV1.GET(GetProductPath, app.deps.handlers.product.GetProduct)
}

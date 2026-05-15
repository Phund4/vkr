package main

import (
	"context"
	"io/fs"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	zlog "github.com/rs/zerolog/log"

	"rtsp-generator/internal/config"
	"rtsp-generator/internal/ffmpeg"
	"rtsp-generator/internal/httpapi"
	"rtsp-generator/internal/logging"
	"rtsp-generator/internal/static"
)

func main() {
	logging.InitGlobal("rtsp-generator")
	cfg := config.Load()
	mgr := ffmpeg.NewManager(cfg.VideoDir, cfg.RTSPPublishBase, cfg.DefaultFPS, cfg.DefaultSize)
	srv := httpapi.New(cfg, mgr)
	e := srv.Echo()
	e.Use(middleware.Recover())

	webRoot := echo.MustSubFS(static.Web, "web")
	indexHTML, err := fs.ReadFile(webRoot, "index.html")
	if err != nil {
		zlog.Fatal().Err(err).Msg("read index.html")
	}
	e.GET("/", func(c echo.Context) error {
		return c.Blob(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	})

	go func() {
		if err := e.Start(cfg.ListenAddr); err != nil && err != http.ErrServerClosed {
			zlog.Error().Err(err).Msg("http server")
		}
	}()

	zlog.Info().
		Str("listen", cfg.ListenAddr).
		Str("rtsp_publish_base", cfg.RTSPPublishBase).
		Str("video_dir", cfg.VideoDir).
		Msg("rtsp-generator UI/API started")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	shCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = e.Shutdown(shCtx)
}

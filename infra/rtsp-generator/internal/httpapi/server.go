package httpapi

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"rtsp-generator/internal/config"
	"rtsp-generator/internal/ffmpeg"
)

// Server HTTP API + статика.
type Server struct {
	e   *echo.Echo
	mgr *ffmpeg.Manager
	cfg config.Config
}

// New создаёт Echo с маршрутами.
func New(cfg config.Config, mgr *ffmpeg.Manager) *Server {
	e := echo.New()
	e.HideBanner = true
	s := &Server{e: e, mgr: mgr, cfg: cfg}

	e.GET("/health", s.health)
	e.GET("/api/config", s.getConfig)
	e.GET("/api/videos", s.listVideos)
	e.GET("/api/streams", s.listStreams)
	e.POST("/api/streams", s.startStream)
	e.DELETE("/api/streams/:id", s.stopStream)

	return s
}

// Echo возвращает экземпляр для embed static на корне.
func (s *Server) Echo() *echo.Echo {
	return s.e
}

func (s *Server) health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) getConfig(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{
		"rtsp_publish_base": s.cfg.RTSPPublishBase,
		"video_dir":         s.cfg.VideoDir,
		"default_fps":       s.cfg.DefaultFPS,
		"default_size":      s.cfg.DefaultSize,
		"hint":              "Укажите в Postgres/coordinator RTSP URL вида {rtsp_publish_base}/{stream_id}",
	})
}

func (s *Server) listVideos(c echo.Context) error {
	v, err := s.mgr.ListVideos()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"videos": v})
}

func (s *Server) listStreams(c echo.Context) error {
	list := s.mgr.List()
	return c.JSON(http.StatusOK, map[string]any{"streams": list})
}

func (s *Server) startStream(c echo.Context) error {
	var spec ffmpeg.Spec
	if err := c.Bind(&spec); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid json")
	}
	run, err := s.mgr.Start(spec)
	if err != nil {
		switch {
		case errors.Is(err, ffmpeg.ErrStreamExists):
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		case errors.Is(err, ffmpeg.ErrInvalidStreamID):
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		case errors.Is(err, ffmpeg.ErrInvalidFile):
			return echo.NewHTTPError(http.StatusBadRequest, "file not found in video dir")
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}
	return c.JSON(http.StatusCreated, run)
}

func (s *Server) stopStream(c echo.Context) error {
	id := c.Param("id")
	if err := s.mgr.Stop(id); err != nil {
		if errors.Is(err, ffmpeg.ErrStreamNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

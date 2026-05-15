package config

import (
	"os"
	"strconv"
	"strings"
)

// Config из переменных окружения.
type Config struct {
	ListenAddr      string
	RTSPPublishBase string
	VideoDir        string
	DefaultFPS      int
	DefaultSize     string
}

// Load читает конфиг с дефолтами для docker-compose.
func Load() Config {
	c := Config{
		ListenAddr:      ":8096",
		RTSPPublishBase: "rtsp://mediamtx:8554",
		VideoDir:        "/videos",
		DefaultFPS:      25,
		DefaultSize:     "1280x720",
	}
	if v := strings.TrimSpace(os.Getenv("LISTEN_ADDR")); v != "" {
		c.ListenAddr = v
	}
	if v := strings.TrimSpace(os.Getenv("RTSP_PUBLISH_BASE")); v != "" {
		c.RTSPPublishBase = strings.TrimRight(v, "/")
	}
	if v := strings.TrimSpace(os.Getenv("SIM_VIDEO_DIR")); v != "" {
		c.VideoDir = v
	}
	if v := strings.TrimSpace(os.Getenv("SIM_FPS")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			c.DefaultFPS = n
		}
	}
	if v := strings.TrimSpace(os.Getenv("SIM_SIZE")); v != "" {
		c.DefaultSize = v
	}
	return c
}

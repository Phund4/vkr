// Package config загружает YAML-конфигурацию и применяет значения по умолчанию.
package config

import (
	"fmt"
	"math"
	"os"

	"gopkg.in/yaml.v3"

	"router/internal/core/domain"
)

// Storage префикс ключей кадров в S3 (загрузка выполняет pusher).
type Storage struct {
	// Prefix необязательный префикс ключей (например its-ingest).
	Prefix string `yaml:"prefix"`
}

// Ingest поведение захвата и выгрузки кадров.
type Ingest struct {
	// TargetFPS целевой FPS для ffmpeg при дискретизации потока.
	TargetFPS float64 `yaml:"target_fps"`

	// FFmpegPath исполняемый файл ffmpeg.
	FFmpegPath string `yaml:"ffmpeg_path"`

	// ProcessWorkers параллельных обработчиков кадра (Kafka + ML) на одну камеру.
	// Если 0 — по умолчанию ceil(target_fps), чтобы скорость обработки могла совпасть с дискретизацией ffmpeg.
	ProcessWorkers int `yaml:"process_workers"`
}

// Metrics экспорт Prometheus.
type Metrics struct {
	// ListenAddr адрес HTTP :port для /metrics.
	ListenAddr string `yaml:"listen_addr"`
}

// Camera один RTSP-источник в YAML (локальный режим без coordinator).
type Camera struct {
	// SegmentID логический сегмент.
	SegmentID string `yaml:"segment_id"`
	// CameraID идентификатор камеры.
	CameraID string `yaml:"camera_id"`
	// RTSPURL URL потока.
	RTSPURL string `yaml:"rtsp_url"`
}

// ToDomain маппинг конфигурации YAML в домен.
func (c Camera) ToDomain() domain.Camera {
	return domain.Camera{
		SegmentID: c.SegmentID,
		CameraID:  c.CameraID,
		RTSPURL:   c.RTSPURL,
	}
}

// Root корневая конфигурация YAML.
type Root struct {
	// Storage префикс ключей для pusher (yaml: storage или legacy s3).
	Storage Storage `yaml:"storage"`

	// Ingest параметры пайплайна кадров.
	Ingest Ingest `yaml:"ingest"`

	// Metrics адрес метрик.
	Metrics Metrics `yaml:"metrics"`

	// Cameras список камер (контур RTSP).
	Cameras []Camera `yaml:"cameras"`

	// ConfigFile путь к загруженному YAML (не из файла).
	ConfigFile string `yaml:"-"`
}

// Load читает YAML по пути path, парсит и валидирует структуру Root.
func Load(path string) (*Root, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var c Root
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("yaml: %w", err)
	}
	c.ConfigFile = path
	ApplyEnvOverrides(&c)
	if err := c.validate(); err != nil {
		return nil, err
	}
	return &c, nil
}

// LoadFromEnv читает CONFIG_PATH или config.cameras.yaml по умолчанию.
func LoadFromEnv() (*Root, error) {
	p := os.Getenv("CONFIG_PATH")
	if p == "" {
		p = "config.cameras.yaml"
	}
	return Load(p)
}

// validate проверяет базовую конфигурацию и подставляет значения по умолчанию.
func (c *Root) validate() error {
	if c.Ingest.TargetFPS <= 0 {
		c.Ingest.TargetFPS = 3
	}
	if c.Ingest.ProcessWorkers <= 0 {
		c.Ingest.ProcessWorkers = int(math.Ceil(c.Ingest.TargetFPS))
	}
	if c.Ingest.ProcessWorkers < 1 {
		c.Ingest.ProcessWorkers = 1
	}
	if c.Ingest.ProcessWorkers > 64 {
		c.Ingest.ProcessWorkers = 64
	}
	if c.Ingest.FFmpegPath == "" {
		c.Ingest.FFmpegPath = "ffmpeg"
	}
	if c.Metrics.ListenAddr == "" {
		c.Metrics.ListenAddr = ":9091"
	}
	return nil
}

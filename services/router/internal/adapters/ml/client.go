package mlclient

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"router/internal/core/domain"
)

// Client HTTP-клиент к одному эндпоинту ML (multipart JPEG + поля метаданных).
type Client struct {
	base        string
	processPath string
	cli         *http.Client
}

// New создаёт клиент с baseURL, относительным processPath и таймаутом HTTP.
func New(baseURL, processPath string, timeout time.Duration) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	return &Client{
		base:        baseURL,
		processPath: processPath,
		cli:         &http.Client{Timeout: timeout},
	}
}

// PostProcess отправляет JPEG в ML; при 204 тело ответа игнорируется.
func (c *Client) PostProcess(ctx context.Context, jpeg []byte, filename string, meta domain.ProcessMeta) error {
	return c.postMultipart(ctx, c.processPath, jpeg, filename, meta)
}

// postMultipart собирает multipart/form-data и выполняет POST.
func (c *Client) postMultipart(ctx context.Context, path string, jpeg []byte, filename string, meta domain.ProcessMeta) error {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile(formFieldImage, filename)
	if err != nil {
		return err
	}
	if _, err := part.Write(jpeg); err != nil {
		return err
	}
	if meta.SegmentID != "" {
		_ = w.WriteField(formFieldSegmentID, meta.SegmentID)
	}
	if meta.CameraID != "" {
		_ = w.WriteField(formFieldCameraID, meta.CameraID)
	}
	if meta.S3Key != "" {
		_ = w.WriteField(formFieldS3Key, meta.S3Key)
	}
	if meta.ObservedAt != "" {
		_ = w.WriteField(formFieldObservedAt, meta.ObservedAt)
	}
	if meta.PipelineStartedAt != "" {
		_ = w.WriteField(formFieldPipelineStartedAt, meta.PipelineStartedAt)
	}
	if err := w.Close(); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+path, &buf)
	if err != nil {
		return err
	}
	req.Header.Set(headerContentType, w.FormDataContentType())
	resp, err := c.cli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, httpErrorBodyMaxBytes))
		return mlHTTPError(resp.StatusCode, strings.TrimSpace(string(b)))
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}

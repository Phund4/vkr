package domain

// FrameWithURL кадр с опциональной presigned ссылкой на S3.
type FrameWithURL struct {
	FrameRef
	DownloadURL string `json:"download_url,omitempty"`
}

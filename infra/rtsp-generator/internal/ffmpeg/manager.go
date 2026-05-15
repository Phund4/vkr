package ffmpeg

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	ErrStreamExists    = errors.New("stream already running")
	ErrStreamNotFound  = errors.New("stream not found")
	ErrInvalidStreamID = errors.New("invalid stream id")
	ErrInvalidFile     = errors.New("invalid or missing video file")
)

// Spec параметры публикации RTSP.
type Spec struct {
	ID            string `json:"stream_id"`
	File          string `json:"file,omitempty"`
	Loop          bool   `json:"loop"`
	Realtime      bool   `json:"realtime"`
	Synthetic     bool   `json:"synthetic"`
	SyntheticKind string `json:"synthetic_kind,omitempty"` // testsrc2 | smptebars
}

// Running описание активного потока.
type Running struct {
	Spec      Spec      `json:"spec"`
	StreamID  string    `json:"stream_id"` // дубль spec.stream_id для простых клиентов/UI
	RTSPURL   string    `json:"rtsp_url"`
	StartedAt time.Time `json:"started_at"`
	LastError string    `json:"last_error,omitempty"`
}

// Manager держит ffmpeg-процессы по stream id.
type Manager struct {
	mu sync.Mutex

	videoDir string
	rtspBase string
	fps      int
	size     string

	procs map[string]*procState
}

type procState struct {
	spec    Spec
	rtspURL string
	started time.Time
	cmd     *exec.Cmd
	cancel  context.CancelFunc
	logBuf  *ringBuffer
}

type ringBuffer struct {
	mu    sync.Mutex
	lines []string
	cap   int
}

func newRing(capacity int) *ringBuffer {
	return &ringBuffer{cap: capacity, lines: make([]string, 0, capacity)}
}

func (r *ringBuffer) add(s string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s = strings.TrimSpace(s)
	if s == "" {
		return
	}
	if len(r.lines) >= r.cap {
		r.lines = r.lines[1:]
	}
	r.lines = append(r.lines, s)
}

func (r *ringBuffer) String() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return strings.Join(r.lines, "\n")
}

// NewManager создаёт менеджер.
func NewManager(videoDir, rtspBase string, fps int, size string) *Manager {
	return &Manager{
		videoDir: filepath.Clean(videoDir),
		rtspBase: strings.TrimRight(rtspBase, "/"),
		fps:      fps,
		size:     size,
		procs:    make(map[string]*procState),
	}
}

// ListVideos возвращает имена .mp4 в каталоге.
func (m *Manager) ListVideos() ([]string, error) {
	entries, err := filepath.Glob(filepath.Join(m.videoDir, "*.mp4"))
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(entries))
	for _, p := range entries {
		out = append(out, filepath.Base(p))
	}
	return out, nil
}

func validateStreamID(id string) bool {
	id = strings.TrimSpace(id)
	if len(id) < 1 || len(id) > 64 {
		return false
	}
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			continue
		}
		return false
	}
	return true
}

func (m *Manager) resolveFile(basename string) (string, error) {
	basename = filepath.Base(basename)
	if basename == "" || basename == "." {
		return "", ErrInvalidFile
	}
	full := filepath.Clean(filepath.Join(m.videoDir, basename))
	rel, err := filepath.Rel(m.videoDir, full)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", ErrInvalidFile
	}
	if _, err := os.Stat(full); err != nil {
		return "", ErrInvalidFile
	}
	return full, nil
}

// Start запускает поток.
// Контекст процесса не привязываем к HTTP-запросу: в net/http контекст запроса
// отменяется сразу после ответа handler, и CommandContext убивает ffmpeg.
func (m *Manager) Start(spec Spec) (*Running, error) {
	if !validateStreamID(spec.ID) {
		return nil, ErrInvalidStreamID
	}
	spec.ID = strings.TrimSpace(spec.ID)
	if spec.Synthetic {
		spec.SyntheticKind = strings.ToLower(strings.TrimSpace(spec.SyntheticKind))
		if spec.SyntheticKind != "smptebars" {
			spec.SyntheticKind = "testsrc2"
		}
	} else {
		if _, err := m.resolveFile(spec.File); err != nil {
			return nil, err
		}
	}

	runCtx, cancel := context.WithCancel(context.Background())
	args, rtspOut, err := m.buildArgs(spec)
	if err != nil {
		cancel()
		return nil, err
	}

	logBuf := newRing(16)
	st := &procState{
		spec:    spec,
		rtspURL: rtspOut,
		started: time.Now().UTC(),
		cancel:  cancel,
		logBuf:  logBuf,
	}

	m.mu.Lock()
	if _, ok := m.procs[spec.ID]; ok {
		m.mu.Unlock()
		cancel()
		return nil, ErrStreamExists
	}
	m.procs[spec.ID] = st
	m.mu.Unlock()

	cmd := exec.CommandContext(runCtx, "ffmpeg", args...)
	cmd.Stderr = writerFunc(func(p []byte) {
		for _, line := range strings.Split(string(p), "\n") {
			logBuf.add(line)
		}
	})

	if err := cmd.Start(); err != nil {
		cancel()
		m.mu.Lock()
		delete(m.procs, spec.ID)
		m.mu.Unlock()
		return nil, fmt.Errorf("ffmpeg start: %w", err)
	}

	m.mu.Lock()
	st.cmd = cmd
	m.mu.Unlock()

	go func() {
		waitErr := cmd.Wait()
		m.mu.Lock()
		delete(m.procs, spec.ID)
		m.mu.Unlock()
		cancel()
		if waitErr != nil && runCtx.Err() == nil {
			logBuf.add("exit: " + waitErr.Error())
		}
	}()

	return &Running{
		Spec:      spec,
		StreamID:  spec.ID,
		RTSPURL:   rtspOut,
		StartedAt: st.started,
	}, nil
}

func (m *Manager) buildArgs(spec Spec) (args []string, rtspURL string, err error) {
	rtspURL = m.rtspBase + "/" + spec.ID
	videoOut := []string{
		"-an",
		"-c:v", "libx264", "-pix_fmt", "yuv420p", "-preset", "veryfast", "-tune", "zerolatency",
		"-f", "rtsp", "-rtsp_transport", "tcp",
		rtspURL,
	}

	if spec.Synthetic {
		kind := spec.SyntheticKind
		var lavfi string
		switch kind {
		case "smptebars":
			lavfi = fmt.Sprintf("smptebars=size=%s:rate=%d", m.size, m.fps)
		default:
			lavfi = fmt.Sprintf("testsrc2=size=%s:rate=%d", m.size, m.fps)
		}
		args = append(args, "-hide_banner", "-loglevel", "warning")
		if spec.Realtime {
			args = append(args, "-re")
		}
		args = append(args, "-f", "lavfi", "-i", lavfi)
		args = append(args, videoOut...)
		return args, rtspURL, nil
	}

	full, err := m.resolveFile(spec.File)
	if err != nil {
		return nil, "", err
	}
	args = append(args, "-hide_banner", "-loglevel", "warning")
	if spec.Realtime {
		args = append(args, "-re")
	}
	if spec.Loop {
		args = append(args, "-stream_loop", "-1")
	}
	args = append(args, "-i", full)
	args = append(args, videoOut...)
	return args, rtspURL, nil
}

type writerFunc func([]byte)

func (w writerFunc) Write(p []byte) (int, error) {
	w(p)
	return len(p), nil
}

// Stop останавливает поток.
func (m *Manager) Stop(id string) error {
	m.mu.Lock()
	st, ok := m.procs[id]
	m.mu.Unlock()
	if !ok {
		return ErrStreamNotFound
	}
	st.cancel()
	if st.cmd != nil && st.cmd.Process != nil {
		_ = st.cmd.Process.Kill()
	}
	return nil
}

// List активные потоки.
func (m *Manager) List() []Running {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Running, 0, len(m.procs))
	for _, st := range m.procs {
		r := Running{
			Spec:      st.spec,
			StreamID:  st.spec.ID,
			RTSPURL:   st.rtspURL,
			StartedAt: st.started,
			LastError: st.logBuf.String(),
		}
		out = append(out, r)
	}
	return out
}

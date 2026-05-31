// Package server serves a local browser UI for configuring and generating
// projects. It is started by `gogen serve` and has no third-party deps.
package server

import (
	"archive/zip"
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/indalyadav56/gogen/internal/blueprint"
	"github.com/indalyadav56/gogen/internal/cli"
	"github.com/indalyadav56/gogen/internal/generator"
	"github.com/indalyadav56/gogen/internal/logx"
)

// DefaultPort is an uncommon port chosen to avoid clashing with typical dev
// services. If it is busy, Serve falls back to a free OS-assigned port.
const DefaultPort = 7720

// Server holds the dependencies needed to render and generate projects.
// templateFS holds the project templates; webFS holds the UI assets (web/).
type Server struct {
	templateFS embed.FS
	webFS      embed.FS
}

// NewServer creates a web server backed by the project templates and UI assets.
func NewServer(templateFS, webFS embed.FS) *Server {
	return &Server{templateFS: templateFS, webFS: webFS}
}

// Serve binds a listener (preferring `port`, else a free port) and serves the
// UI until the process is stopped. It opens the browser on a best-effort basis.
func Serve(templateFS, webFS embed.FS, port int) error {
	ln, err := listen(port)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("http://127.0.0.1:%d", ln.Addr().(*net.TCPAddr).Port)

	s := NewServer(templateFS, webFS)
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/options", s.handleOptions)
	mux.HandleFunc("/api/preview", s.handlePreview)
	mux.HandleFunc("/api/generate", s.handleGenerate)

	fmt.Printf("🚀 gogen UI running at %s\n   press Ctrl+C to stop\n", url)
	logx.S().Infow("serving UI", "url", url)
	openBrowser(url)
	return http.Serve(ln, withLogging(mux))
}

// statusRecorder captures the response status for request logging.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// withLogging logs every request's method, path, status and duration.
func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		logx.S().Infow("request",
			"method", r.Method, "path", r.URL.Path,
			"status", rec.status, "dur", time.Since(start).String())
	})
}

// listen prefers the requested port but falls back to a free ephemeral port so
// it never collides with another service on the user's machine.
func listen(port int) (net.Listener, error) {
	if ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port)); err == nil {
		return ln, nil
	}
	return net.Listen("tcp", "127.0.0.1:0")
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	page, err := s.webFS.ReadFile("web/index.html")
	if err != nil {
		http.Error(w, "UI asset not found", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(page)
}

func (s *Server) handleOptions(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"architectures": cli.Architectures,
		"routers":       cli.Routers,
		"databases":     cli.Databases,
		"frontends":     cli.Frontends,
	})
}

// request is the JSON payload from the UI for preview and generate.
type request struct {
	Module   string   `json:"module"`
	Arch     string   `json:"arch"`
	Router   string   `json:"router"`
	DB       string   `json:"db"`
	Entities []string `json:"entities"`
	Auth     bool     `json:"auth"`
	Frontend string   `json:"frontend"`
}

func (req request) toConfig() *cli.Config {
	entities := make([]string, 0, len(req.Entities))
	for _, e := range req.Entities {
		if e = strings.TrimSpace(e); e != "" {
			entities = append(entities, e)
		}
	}
	if len(entities) == 0 {
		entities = []string{"Item"}
	}
	return &cli.Config{
		ModuleName: strings.TrimSpace(req.Module),
		Arch:       strings.ToLower(strings.TrimSpace(req.Arch)),
		Router:     strings.ToLower(strings.TrimSpace(req.Router)),
		DB:         strings.ToLower(strings.TrimSpace(req.DB)),
		Entities:   entities,
		Auth:       req.Auth,
		Frontend:   strings.ToLower(strings.TrimSpace(req.Frontend)),
	}
}

// handlePreview returns the sorted list of files the chosen config would create.
func (s *Server) handlePreview(w http.ResponseWriter, r *http.Request) {
	cfg, err := decodeConfig(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	bp, ok := blueprint.Get(cfg.Arch)
	if !ok {
		writeError(w, http.StatusBadRequest, fmt.Errorf("unknown architecture %q", cfg.Arch))
		return
	}
	paths := make([]string, 0)
	for _, f := range bp.Build(cfg) {
		paths = append(paths, f.Path)
	}
	sort.Strings(paths)
	writeJSON(w, http.StatusOK, map[string]any{"project": cfg.GetProjectRoot(), "files": paths})
}

// handleGenerate scaffolds the project into a temp dir and streams it as a zip.
func (s *Server) handleGenerate(w http.ResponseWriter, r *http.Request) {
	cfg, err := decodeConfig(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	tmp, err := os.MkdirTemp("", "gogen-")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	defer os.RemoveAll(tmp)

	g := generator.NewProjectGeneratorInDir(cfg, s.templateFS, tmp)
	if err := g.Scaffold(); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	// Write a go.mod directly (no toolchain/network); the README tells the user
	// to run `go mod tidy` after extracting.
	gomod := fmt.Sprintf("module %s\n\ngo 1.24\n", cfg.ModuleName)
	if err := os.WriteFile(filepath.Join(g.ProjectDir(), "go.mod"), []byte(gomod), 0o644); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	archive, err := zipDir(g.ProjectDir())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", cfg.GetProjectRoot()+".zip"))
	_, _ = w.Write(archive)
}

func decodeConfig(r *http.Request) (*cli.Config, error) {
	if r.Method != http.MethodPost {
		return nil, fmt.Errorf("method not allowed")
	}
	var req request
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil {
		return nil, fmt.Errorf("invalid request body: %w", err)
	}
	cfg := req.toConfig()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// zipDir packs the directory at root into an in-memory zip, storing entries
// relative to root's parent (so the project folder is the top-level entry).
func zipDir(root string) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	base := filepath.Dir(root)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, err := filepath.Rel(base, path)
		if err != nil {
			return err
		}
		fw, err := zw.Create(filepath.ToSlash(rel))
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		_, err = fw.Write(data)
		return err
	})
	if err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	logx.S().Warnw("request error", "status", status, "error", err)
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

// openBrowser tries to open url in the default browser; failures are ignored.
func openBrowser(url string) {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd, args = "rundll32", []string{"url.dll,FileProtocolHandler"}
	default:
		cmd = "xdg-open"
	}
	_ = exec.Command(cmd, append(args, url)...).Start()
}

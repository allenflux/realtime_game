package main

import (
	"embed"
	"encoding/json"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

// 通过 embed 打包模板和静态资源，部署时只需要一个二进制文件。

//go:embed templates/* static/*
var assets embed.FS

type Config struct {
	ListenAddr string
	BackendURL string
}

type App struct {
	cfg       Config
	templates *template.Template
	client    *http.Client
}

type APIError struct {
	Message string `json:"message"`
}

func main() {
	cfg := Config{
		ListenAddr: getenv("FRONTEND_LISTEN_ADDR", ":8090"),
		BackendURL: strings.TrimRight(getenv("GAME_BACKEND_URL", "http://127.0.0.1:18080"), "/"),
	}

	tmpl := template.Must(template.ParseFS(assets, "templates/*.html"))

	staticFS, err := fs.Sub(assets, "static")
	if err != nil {
		log.Fatalf("init static fs failed: %v", err)
	}

	app := &App{
		cfg:       cfg,
		templates: tmpl,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", app.indexHandler)
	mux.HandleFunc("/api/config", app.configHandler)
	mux.HandleFunc("/api/proxy/current-round", app.proxyGet)
	mux.HandleFunc("/api/proxy/leaderboard", app.proxyGet)
	mux.HandleFunc("/api/proxy/jackpot", app.proxyGet)
	mux.HandleFunc("/api/proxy/my-bets", app.proxyGet)
	mux.HandleFunc("/api/proxy/bet", app.proxyPost)
	mux.HandleFunc("/api/proxy/cashout", app.proxyPost)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	server := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           loggingMiddleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("frontend listen on %s, backend=%s", cfg.ListenAddr, cfg.BackendURL)
	log.Fatal(server.ListenAndServe())
}

func (a *App) indexHandler(w http.ResponseWriter, r *http.Request) {
	channelID, _ := strconv.ParseInt(r.URL.Query().Get("channel_id"), 10, 64)
	if channelID <= 0 {
		channelID = 1001
	}
	token := strings.TrimSpace(r.URL.Query().Get("api_sys_token"))
	if token == "" {
		token = "token-demo-1"
	}

	err := a.templates.ExecuteTemplate(w, "index.html", map[string]any{
		"ChannelID":   channelID,
		"ApiSysToken": token,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *App) configHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"backend_url": a.cfg.BackendURL,
		"server_time": time.Now().UnixMilli(),
	})
}

func (a *App) proxyGet(w http.ResponseWriter, r *http.Request) {
	endpoint := mapProxyPath(r.URL.Path)
	if endpoint == "" {
		writeJSON(w, http.StatusNotFound, APIError{Message: "unknown api path"})
		return
	}

	target := a.cfg.BackendURL + endpoint
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, target, nil)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Message: err.Error()})
		return
	}

	resp, err := a.client.Do(req)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, APIError{Message: err.Error()})
		return
	}
	defer resp.Body.Close()

	copyResponse(w, resp)
}

func (a *App) proxyPost(w http.ResponseWriter, r *http.Request) {
	endpoint := mapProxyPath(r.URL.Path)
	if endpoint == "" {
		writeJSON(w, http.StatusNotFound, APIError{Message: "unknown api path"})
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIError{Message: err.Error()})
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, a.cfg.BackendURL+endpoint, strings.NewReader(string(body)))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, APIError{Message: err.Error()})
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, APIError{Message: err.Error()})
		return
	}
	defer resp.Body.Close()

	copyResponse(w, resp)
}

func mapProxyPath(p string) string {
	switch path.Clean(p) {
	case "/api/proxy/current-round":
		return "/v2/game/current-round"
	case "/api/proxy/profile":
		return "/v2/game/profile"
	case "/api/proxy/leaderboard":
		return "/v2/game/leaderboard"
	case "/api/proxy/jackpot":
		return "/v2/game/jackpot"
	case "/api/proxy/my-bets":
		return "/v2/game/my-bets"
	case "/api/proxy/bet":
		return "/v2/game/bet"
	case "/api/proxy/cashout":
		return "/v2/game/cashout"
	default:
		return ""
	}
}

func copyResponse(w http.ResponseWriter, resp *http.Response) {
	for k, values := range resp.Header {
		if strings.EqualFold(k, "Content-Length") {
			continue
		}
		for _, v := range values {
			w.Header().Add(k, v)
		}
	}
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		begin := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s cost=%s", r.Method, r.URL.Path, time.Since(begin))
	})
}

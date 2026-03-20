package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/devstack/devstack/internal/config"
	"github.com/devstack/devstack/internal/services/redis"
	"github.com/devstack/devstack/internal/services/s3"
	"github.com/devstack/devstack/internal/services/smtp"
)

type Server struct {
	mu      sync.RWMutex
	cfg     *config.DashboardConfig
	redis   *redis.Service
	s3      *s3.Service
	smtp    *smtp.Service
	running bool
	httpSrv *http.Server
	logs    []string
	logMu   sync.RWMutex
	logFunc func(string)
}

func New(cfg *config.DashboardConfig, r *redis.Service, s *s3.Service, sm *smtp.Service) *Server {
	return &Server{
		cfg:   cfg,
		redis: r,
		s3:    s,
		smtp:  sm,
	}
}

func (s *Server) SetLogFunc(f func(string)) {
	s.logFunc = f
}

func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	mux := http.NewServeMux()

	fsys := GetAssets()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "" || path == "/" {
			path = "index.html"
		}
		if len(path) > 0 && path[0] == '/' {
			path = path[1:]
		}

		content, err := fs.ReadFile(fsys, path)
		if err == nil {
			ext := filepath.Ext(path)
			contentType := "text/plain"
			switch ext {
			case ".html":
				contentType = "text/html"
			case ".css":
				contentType = "text/css"
			case ".js":
				contentType = "application/javascript"
			case ".json":
				contentType = "application/json"
			case ".svg":
				contentType = "image/svg+xml"
			case ".png":
				contentType = "image/png"
			case ".ico":
				contentType = "image/x-icon"
			}
			w.Header().Set("Content-Type", contentType)
			w.Write(content)
			return
		}

		content, err = fs.ReadFile(fsys, "index.html")
		if err != nil {
			http.Error(w, "Frontend not embedded. Run: cd frontend && npm install && npm run build", 500)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write(content)
	})

	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/redis/", s.handleRedisAPI)
	mux.HandleFunc("/api/s3/", s.handleS3API)
	mux.HandleFunc("/api/smtp/", s.handleSMTPAPI)
	mux.HandleFunc("/api/persist", s.handlePersist)
	mux.HandleFunc("/api/logs", s.handleLogs)

	s.httpSrv = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port),
		Handler: mux,
	}

	s.running = true

	go func() {
		if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Dashboard server error: %v", err)
		}
	}()

	s.addLog("Dashboard started on %s:%d", s.cfg.Host, s.cfg.Port)
	if s.logFunc != nil {
		s.logFunc(fmt.Sprintf("Dashboard started on %s:%d", s.cfg.Host, s.cfg.Port))
	}
	log.Printf("Dashboard: serving embedded React app")
	log.Printf("Dashboard: started on %s:%d", s.cfg.Host, s.cfg.Port)
	return nil
}

func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.httpSrv.Shutdown(ctx)
	s.running = false
	s.addLog("Dashboard stopped")
	if s.logFunc != nil {
		s.logFunc("Dashboard stopped")
	}
	log.Println("Dashboard: stopped")
	return nil
}

func (s *Server) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

func (s *Server) addLog(format string, args ...interface{}) {
	s.logMu.Lock()
	defer s.logMu.Unlock()
	msg := fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), fmt.Sprintf(format, args...))
	s.logs = append(s.logs, msg)
	if len(s.logs) > 500 {
		s.logs = s.logs[len(s.logs)-500:]
	}
	if s.logFunc != nil {
		s.logFunc(msg)
	}
}

func JsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	JsonResponse(w, map[string]string{
		"redis":     s.boolToStatus(s.redis.IsRunning()),
		"s3":        s.boolToStatus(s.s3.IsRunning()),
		"smtp":      s.boolToStatus(s.smtp.IsRunning()),
		"dashboard": s.boolToStatus(s.IsRunning()),
	})
}

func (s *Server) boolToStatus(b bool) string {
	if b {
		return "running"
	}
	return "stopped"
}

func (s *Server) handleRedisAPI(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/api/redis/"):]

	switch r.Method {
	case "GET":
		if path == "keys" {
			keys, err := s.redis.GetKeys()
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			JsonResponse(w, keys)
			return
		}
		if len(path) > 0 && path[len(path)-1] != '/' {
			key := filepath.Base(path)
			val, err := s.redis.GetKey(key)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			JsonResponse(w, map[string]string{"key": key, "value": val})
			return
		}
		stats, _ := s.redis.GetStats()
		JsonResponse(w, stats)
	case "POST":
		switch path {
		case "start":
			s.redis.Start()
			s.addLog("Redis started")
			if s.logFunc != nil {
				s.logFunc("Redis started")
			}
			w.Write([]byte(`{"status":"ok"}`))
		case "stop":
			s.redis.Stop()
			s.addLog("Redis stopped")
			if s.logFunc != nil {
				s.logFunc("Redis stopped")
			}
			w.Write([]byte(`{"status":"ok"}`))
		case "restart":
			s.redis.Restart()
			s.addLog("Redis restarted")
			if s.logFunc != nil {
				s.logFunc("Redis restarted")
			}
			w.Write([]byte(`{"status":"ok"}`))
		case "save":
			s.redis.Save()
			s.addLog("Redis data saved")
			if s.logFunc != nil {
				s.logFunc("Redis data saved")
			}
			w.Write([]byte(`{"status":"ok"}`))
		}
	case "DELETE":
		key := filepath.Base(path)
		if err := s.redis.DeleteKey(key); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		s.addLog("Redis key deleted: %s", key)
		if s.logFunc != nil {
			s.logFunc(fmt.Sprintf("Redis key deleted: %s", key))
		}
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func (s *Server) handleS3API(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/api/s3/"):]

	switch r.Method {
	case "GET":
		if path == "buckets" {
			buckets, err := s.s3.ListBuckets()
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			JsonResponse(w, buckets)
			return
		}
		parts := strings.SplitN(path, "/", 4)
		if len(parts) >= 3 && parts[1] == "bucket" && len(parts) == 4 && parts[3] == "objects" {
			bucket := parts[2]
			objects, err := s.s3.ListObjects(bucket)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			JsonResponse(w, objects)
			return
		}
		stats, _ := s.s3.GetStats()
		JsonResponse(w, stats)
	case "POST":
		switch path {
		case "start":
			s.s3.Start()
			s.addLog("S3 started")
			if s.logFunc != nil {
				s.logFunc("S3 started")
			}
			w.Write([]byte(`{"status":"ok"}`))
		case "stop":
			s.s3.Stop()
			s.addLog("S3 stopped")
			if s.logFunc != nil {
				s.logFunc("S3 stopped")
			}
			w.Write([]byte(`{"status":"ok"}`))
		case "restart":
			s.s3.Restart()
			s.addLog("S3 restarted")
			if s.logFunc != nil {
				s.logFunc("S3 restarted")
			}
			w.Write([]byte(`{"status":"ok"}`))
		}
		if strings.HasPrefix(path, "bucket/") {
			bucket := strings.TrimPrefix(path, "bucket/")
			if err := s.s3.CreateBucket(bucket); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			s.addLog("S3 bucket created: %s", bucket)
			if s.logFunc != nil {
				s.logFunc(fmt.Sprintf("S3 bucket created: %s", bucket))
			}
			w.Write([]byte(`{"status":"ok"}`))
		}
	case "PUT":
		if strings.HasPrefix(path, "bucket/") {
			bucket := strings.TrimPrefix(path, "bucket/")
			if err := s.s3.CreateBucket(bucket); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			s.addLog("S3 bucket created: %s", bucket)
			if s.logFunc != nil {
				s.logFunc(fmt.Sprintf("S3 bucket created: %s", bucket))
			}
			w.Write([]byte(`{"status":"ok"}`))
		}
	case "DELETE":
		if strings.HasPrefix(path, "bucket/") {
			parts := strings.SplitN(path, "/", 3)
			if len(parts) == 3 && parts[2] == "" {
				bucket := parts[1]
				if err := s.s3.DeleteBucket(bucket); err != nil {
					http.Error(w, err.Error(), 500)
					return
				}
				s.addLog("S3 bucket deleted: %s", bucket)
				if s.logFunc != nil {
					s.logFunc(fmt.Sprintf("S3 bucket deleted: %s", bucket))
				}
				w.Write([]byte(`{"status":"ok"}`))
			}
		}
		if strings.Contains(path, "/object/") {
			parts := strings.SplitN(path, "/object/", 2)
			if len(parts) == 2 {
				bucket, key := parts[0], parts[1]
				if err := s.s3.DeleteObject(bucket, key); err != nil {
					http.Error(w, err.Error(), 500)
					return
				}
				s.addLog("S3 object deleted: %s/%s", bucket, key)
				if s.logFunc != nil {
					s.logFunc(fmt.Sprintf("S3 object deleted: %s/%s", bucket, key))
				}
				w.Write([]byte(`{"status":"ok"}`))
			}
		}
	}
}

func (s *Server) handleSMTPAPI(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/api/smtp/"):]

	switch r.Method {
	case "GET":
		if path == "emails" {
			emails := s.smtp.ListEmails()
			JsonResponse(w, emails)
			return
		}
		if len(path) > 0 {
			id := filepath.Base(path)
			email := s.smtp.GetEmail(id)
			if email == nil {
				http.Error(w, "email not found", 404)
				return
			}
			JsonResponse(w, email)
			return
		}
		stats := s.smtp.GetStats()
		JsonResponse(w, stats)
	case "POST":
		switch path {
		case "start":
			s.smtp.Start()
			s.addLog("SMTP started")
			if s.logFunc != nil {
				s.logFunc("SMTP started")
			}
			w.Write([]byte(`{"status":"ok"}`))
		case "stop":
			s.smtp.Stop()
			s.addLog("SMTP stopped")
			if s.logFunc != nil {
				s.logFunc("SMTP stopped")
			}
			w.Write([]byte(`{"status":"ok"}`))
		case "restart":
			s.smtp.Restart()
			s.addLog("SMTP restarted")
			if s.logFunc != nil {
				s.logFunc("SMTP restarted")
			}
			w.Write([]byte(`{"status":"ok"}`))
		case "clear":
			s.smtp.ClearEmails()
			s.addLog("SMTP emails cleared")
			if s.logFunc != nil {
				s.logFunc("SMTP emails cleared")
			}
			w.Write([]byte(`{"status":"ok"}`))
		}
	case "DELETE":
		id := filepath.Base(path)
		if err := s.smtp.DeleteEmail(id); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		s.addLog("SMTP email deleted: %s", id)
		if s.logFunc != nil {
			s.logFunc(fmt.Sprintf("SMTP email deleted: %s", id))
		}
		w.Write([]byte(`{"status":"ok"}`))
	}
}

func (s *Server) handlePersist(w http.ResponseWriter, r *http.Request) {
	s.redis.Save()
	s.addLog("All data persisted")
	if s.logFunc != nil {
		s.logFunc("All data persisted")
	}
	w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	s.logMu.RLock()
	defer s.logMu.RUnlock()
	JsonResponse(w, s.logs)
}

func (s *Server) GetAddr() string {
	return fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
}

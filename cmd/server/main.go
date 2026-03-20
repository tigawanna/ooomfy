package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/devstack/devstack/internal/config"
	"github.com/devstack/devstack/internal/dashboard"
	"github.com/devstack/devstack/internal/services/redis"
	"github.com/devstack/devstack/internal/services/s3"
	"github.com/devstack/devstack/internal/services/smtp"
)

var (
	flagReset     bool
	flagConfig    string
	flagNoPersist bool
)

var (
	redisPort, s3Port, smtpPort, dashboardPort int
	redisHost, s3Host, smtpHost, dashboardHost string
)

var (
	logChan = make(chan string, 100)
	logs    []string
	logsMu  sync.RWMutex
)

func init() {
	flag.BoolVar(&flagReset, "reset", false, "Clear all persisted data before starting")
	flag.StringVar(&flagConfig, "config", "", "Path to config file")
	flag.BoolVar(&flagNoPersist, "no-persist", false, "Disable data persistence")

	flag.IntVar(&redisPort, "redis-port", 0, "Redis port (0 = use config)")
	flag.IntVar(&s3Port, "s3-port", 0, "S3 port (0 = use config)")
	flag.IntVar(&smtpPort, "smtp-port", 0, "SMTP port (0 = use config)")
	flag.IntVar(&dashboardPort, "dashboard-port", 0, "Dashboard port (0 = use config)")

	flag.StringVar(&redisHost, "redis-host", "", "Redis host (empty = use config)")
	flag.StringVar(&s3Host, "s3-host", "", "S3 host (empty = use config)")
	flag.StringVar(&smtpHost, "smtp-host", "", "SMTP host (empty = use config)")
	flag.StringVar(&dashboardHost, "dashboard-host", "", "Dashboard host (empty = use config)")
}

type logWriter struct{}

func (l logWriter) Write(p []byte) (n int, err error) {
	s := string(p)
	logChan <- s
	addLog(s)
	return len(p), nil
}

func addLog(s string) {
	logsMu.Lock()
	defer logsMu.Unlock()
	logs = append(logs, s)
	if len(logs) > 500 {
		logs = logs[len(logs)-500:]
	}
}

func getLogs() []string {
	logsMu.RLock()
	defer logsMu.RUnlock()
	result := make([]string, len(logs))
	copy(result, logs)
	return result
}

func main() {
	log.SetOutput(logWriter{})

	flag.Parse()

	cfg, err := config.Load(flagConfig)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if flagNoPersist {
		cfg.Persist.Enabled = false
	}

	if flagReset {
		reset(cfg)
	}

	if redisPort > 0 {
		cfg.Redis.Port = redisPort
	}
	if redisHost != "" {
		cfg.Redis.Host = redisHost
	}
	if s3Port > 0 {
		cfg.S3.Port = s3Port
	}
	if s3Host != "" {
		cfg.S3.Host = s3Host
	}
	if smtpPort > 0 {
		cfg.SMTP.Port = smtpPort
	}
	if smtpHost != "" {
		cfg.SMTP.Host = smtpHost
	}
	if dashboardPort > 0 {
		cfg.Dashboard.Port = dashboardPort
	}
	if dashboardHost != "" {
		cfg.Dashboard.Host = dashboardHost
	}

	redisSvc := redis.New(&cfg.Redis, &cfg.Persist.Redis)
	s3Svc := s3.New(&cfg.S3, &cfg.Persist.S3)
	smtpSvc := smtp.New(&cfg.SMTP, &cfg.Persist.SMTP)
	dashboardSvc := dashboard.New(&cfg.Dashboard, redisSvc, s3Svc, smtpSvc)
	dashboardSvc.SetLogFunc(addLog)

	log.Println("Starting DevStack Manager...")
	log.Printf("Redis:    %s:%d", cfg.Redis.Host, cfg.Redis.Port)
	log.Printf("S3:       %s:%d", cfg.S3.Host, cfg.S3.Port)
	log.Printf("SMTP:     %s:%d", cfg.SMTP.Host, cfg.SMTP.Port)
	log.Printf("Dashboard: %s:%d", cfg.Dashboard.Host, cfg.Dashboard.Port)

	if err := redisSvc.Start(); err != nil {
		log.Printf("Failed to start Redis: %v", err)
	} else {
		log.Println("Redis started successfully")
	}
	if err := s3Svc.Start(); err != nil {
		log.Printf("Failed to start S3: %v", err)
	} else {
		log.Println("S3 started successfully")
	}
	if err := smtpSvc.Start(); err != nil {
		log.Printf("Failed to start SMTP: %v", err)
	} else {
		log.Println("SMTP started successfully")
	}
	if err := dashboardSvc.Start(); err != nil {
		log.Fatalf("Failed to start dashboard: %v", err)
	}

	printStatus(cfg, redisSvc, s3Svc, smtpSvc, dashboardSvc)

	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/api/logs/stream", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")
			flusher, ok := w.(http.Flusher)
			if !ok {
				return
			}

			clientGone := r.Context().Done()
			for {
				select {
				case <-clientGone:
					return
				case logEntry := <-logChan:
					fmt.Fprintf(w, "data: %s\n\n", logEntry)
					flusher.Flush()
				}
			}
		})
		mux.HandleFunc("/api/logs/all", func(w http.ResponseWriter, r *http.Request) {
			dashboard.JsonResponse(w, getLogs())
		})
		port := cfg.Dashboard.Port + 1
		if port == 65535 {
			port = 9090
		}
		http.ListenAndServe(fmt.Sprintf("%s:%d", cfg.Dashboard.Host, port), mux)
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("\nShutting down...")

	dashboardSvc.Stop()
	redisSvc.Stop()
	s3Svc.Stop()
	smtpSvc.Stop()

	log.Println("OOOMFS stopped")
}

func printStatus(cfg *config.Config, redisSvc *redis.Service, s3Svc *s3.Service, smtpSvc *smtp.Service, dashboardSvc *dashboard.Server) {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                         OOOMFS                             ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════╣")
	fmt.Println("║  Service    Status    Address                            ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════╣")

	redisStatus := "stopped"
	if redisSvc.IsRunning() {
		redisStatus = "running  "
	}
	s3Status := "stopped"
	if s3Svc.IsRunning() {
		s3Status = "running  "
	}
	smtpStatus := "stopped"
	if smtpSvc.IsRunning() {
		smtpStatus = "running  "
	}

	fmt.Printf("║  Redis      %s  %s:%d                      ║\n", redisStatus, cfg.Redis.Host, cfg.Redis.Port)
	fmt.Printf("║  S3         %s  %s:%d                       ║\n", s3Status, cfg.S3.Host, cfg.S3.Port)
	fmt.Printf("║  SMTP       %s  %s:%d                       ║\n", smtpStatus, cfg.SMTP.Host, cfg.SMTP.Port)
	fmt.Printf("║  Dashboard  running    http://%s:%d                   ║\n", cfg.Dashboard.Host, cfg.Dashboard.Port)
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop")
}

func reset(cfg *config.Config) {
	log.Println("Resetting all data...")
	os.RemoveAll(cfg.Persist.Dir)
	log.Println("All data cleared. Starting fresh.")
}

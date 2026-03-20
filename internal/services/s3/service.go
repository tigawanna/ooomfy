package s3

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/devstack/devstack/internal/config"
	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
)

type Service struct {
	mu      sync.RWMutex
	server  *gofakes3.GoFakeS3
	backend *s3mem.Backend
	cfg     *config.S3Config
	persist *config.S3PersistConfig
	running bool
	httpSrv *http.Server
}

func New(cfg *config.S3Config, persist *config.S3PersistConfig) *Service {
	return &Service{
		cfg:     cfg,
		persist: persist,
		running: false,
	}
}

func isPortAvailable(host string, port int) bool {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func findAvailablePort(host string, startPort int, maxAttempts int) int {
	for i := 0; i < maxAttempts; i++ {
		port := startPort + i
		if isPortAvailable(host, port) {
			return port
		}
		log.Printf("S3: port %d already in use, trying %d", port, port+1)
	}
	return startPort
}

func (s *Service) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	port := findAvailablePort(s.cfg.Host, s.cfg.Port, 100)
	if port != s.cfg.Port {
		s.cfg.Port = port
		log.Printf("S3: requested port unavailable, using port %d instead", port)
	}

	if err := os.MkdirAll(s.persist.Directory, 0755); err != nil {
		return fmt.Errorf("failed to create S3 data directory: %w", err)
	}

	backend := s3mem.New()
	s.backend = backend

	server := gofakes3.New(backend)
	s.server = server

	httpSrv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port),
		Handler: s.server.Server(),
	}

	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("S3 server error: %v", err)
		}
	}()

	s.httpSrv = httpSrv
	s.running = true

	log.Printf("S3: started on %s:%d", s.cfg.Host, s.cfg.Port)
	return nil
}

func (s *Service) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if s.httpSrv != nil {
		s.httpSrv.Shutdown(ctx)
	}

	s.running = false
	s.server = nil
	s.backend = nil
	s.httpSrv = nil

	log.Println("S3: stopped")
	return nil
}

func (s *Service) Restart() error {
	if err := s.Stop(); err != nil {
		return err
	}
	return s.Start()
}

func (s *Service) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

func (s *Service) GetPort() int {
	return s.cfg.Port
}

func (s *Service) GetHost() string {
	return s.cfg.Host
}

func (s *Service) GetAddr() string {
	return fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
}

func (s *Service) GetDataDir() string {
	return s.persist.Directory
}

type BucketInfo struct {
	Name         string    `json:"name"`
	CreationDate time.Time `json:"creationDate"`
}

func (s *Service) ListBuckets() ([]BucketInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.running || s.backend == nil {
		return nil, fmt.Errorf("s3 not running")
	}

	buckets, err := s.backend.ListBuckets()
	if err != nil {
		return nil, err
	}

	var result []BucketInfo
	for _, b := range buckets {
		result = append(result, BucketInfo{
			Name:         b.Name,
			CreationDate: b.CreationDate.Time,
		})
	}

	return result, nil
}

func (s *Service) CreateBucket(name string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.running || s.backend == nil {
		return fmt.Errorf("s3 not running")
	}

	return s.backend.CreateBucket(name)
}

func (s *Service) DeleteBucket(name string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.running || s.backend == nil {
		return fmt.Errorf("s3 not running")
	}

	return s.backend.DeleteBucket(name)
}

type ObjectInfo struct {
	Key  string `json:"key"`
	Size int64  `json:"size"`
}

func (s *Service) ListObjects(bucket string) ([]ObjectInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.running || s.backend == nil {
		return nil, fmt.Errorf("s3 not running")
	}

	result, err := s.backend.ListBucket(bucket, nil, gofakes3.ListBucketPage{})
	if err != nil {
		return nil, err
	}

	var objects []ObjectInfo
	for _, item := range result.Contents {
		objects = append(objects, ObjectInfo{
			Key:  item.Key,
			Size: item.Size,
		})
	}

	return objects, nil
}

func (s *Service) DeleteObject(bucket, key string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.running || s.backend == nil {
		return fmt.Errorf("s3 not running")
	}

	_, err := s.backend.DeleteObject(bucket, key)
	return err
}

type Stats struct {
	NumBuckets int   `json:"numBuckets"`
	NumObjects int   `json:"numObjects"`
	TotalSize  int64 `json:"totalSize"`
}

func (s *Service) GetStats() (*Stats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.running {
		return nil, fmt.Errorf("s3 not running")
	}

	return &Stats{
		NumBuckets: 0,
		NumObjects: 0,
		TotalSize:  0,
	}, nil
}

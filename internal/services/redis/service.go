package redis

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"github.com/alicebob/miniredis/v2"
	"github.com/devstack/devstack/internal/config"
)

type Service struct {
	mu       sync.RWMutex
	mr       *miniredis.Miniredis
	cfg      *config.RedisConfig
	persist  *config.RedisPersistConfig
	running  bool
	listener net.Listener
}

func New(cfg *config.RedisConfig, persist *config.RedisPersistConfig) *Service {
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
		log.Printf("Redis: port %d already in use, trying %d", port, port+1)
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
		log.Printf("Redis: requested port unavailable, using port %d instead", port)
	}

	mr := miniredis.NewMiniRedis()
	s.mr = mr

	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	if err := mr.StartAddr(addr); err != nil {
		return fmt.Errorf("failed to start miniredis on %s: %w", addr, err)
	}

	s.running = true

	if s.persist.Enabled {
		if _, err := os.ReadFile(s.persist.File); err == nil {
			log.Printf("Redis: persistence loading not supported in this version")
		}
	}

	log.Printf("Redis: started on %s:%d", s.cfg.Host, s.cfg.Port)
	return nil
}

func (s *Service) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	if s.persist.Enabled {
		s.save()
	}

	s.mr.Close()
	s.running = false
	s.mr = nil

	log.Println("Redis: stopped")
	return nil
}

func (s *Service) Restart() error {
	if err := s.Stop(); err != nil {
		return err
	}
	return s.Start()
}

func (s *Service) save() error {
	if !s.persist.Enabled || s.mr == nil {
		return nil
	}

	dump := s.mr.Dump()

	if err := os.MkdirAll(s.persist.File[:len(s.persist.File)-len("/redis.dump")], 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	tmpFile := s.persist.File + ".tmp"
	if err := os.WriteFile(tmpFile, []byte(dump), 0644); err != nil {
		return fmt.Errorf("failed to write dump file: %w", err)
	}

	if err := os.Rename(tmpFile, s.persist.File); err != nil {
		return fmt.Errorf("failed to rename dump file: %w", err)
	}

	log.Println("Redis: saved data to", s.persist.File)
	return nil
}

func (s *Service) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

func (s *Service) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.save()
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

type KeyInfo struct {
	Key  string `json:"key"`
	Type string `json:"type"`
}

func (s *Service) GetKeys() ([]KeyInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.running || s.mr == nil {
		return nil, fmt.Errorf("redis not running")
	}

	keys := s.mr.Keys()
	result := make([]KeyInfo, 0, len(keys))

	for _, key := range keys {
		keyType := s.mr.Type(key)
		result = append(result, KeyInfo{
			Key:  key,
			Type: keyType,
		})
	}

	return result, nil
}

func (s *Service) GetKey(key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.running || s.mr == nil {
		return "", fmt.Errorf("redis not running")
	}

	keyType := s.mr.Type(key)

	switch keyType {
	case "string":
		val, _ := s.mr.Get(key)
		return val, nil
	case "list":
		items, _ := s.mr.List(key)
		result := "["
		for i, item := range items {
			if i > 0 {
				result += ", "
			}
			result += fmt.Sprintf("%q", item)
		}
		result += "]"
		return result, nil
	case "hash":
		keys, _ := s.mr.HKeys(key)
		result := "{"
		for i, k := range keys {
			if i > 0 {
				result += ", "
			}
			result += fmt.Sprintf("%q: ...", k)
		}
		result += "}"
		return result, nil
	case "set":
		members, _ := s.mr.Members(key)
		result := "["
		for i, m := range members {
			if i > 0 {
				result += ", "
			}
			result += fmt.Sprintf("%q", m)
		}
		result += "]"
		return result, nil
	case "zset":
		set, _ := s.mr.SortedSet(key)
		result := "["
		i := 0
		for k, v := range set {
			if i > 0 {
				result += ", "
			}
			result += fmt.Sprintf("{score: %f, member: %q}", v, k)
			i++
		}
		result += "]"
		return result, nil
	default:
		return "", fmt.Errorf("unsupported type: %s", keyType)
	}
}

func (s *Service) DeleteKey(key string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.running || s.mr == nil {
		return fmt.Errorf("redis not running")
	}

	s.mr.Del(key)
	return nil
}

type Stats struct {
	NumKeys     int    `json:"numKeys"`
	NumCommands int64  `json:"numCommands"`
	Uptime      string `json:"uptime"`
	Version     string `json:"version"`
}

func (s *Service) GetStats() (*Stats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.running || s.mr == nil {
		return nil, fmt.Errorf("redis not running")
	}

	keys := s.mr.Keys()

	return &Stats{
		NumKeys:     len(keys),
		NumCommands: int64(s.mr.CommandCount()),
		Uptime:      "0s",
		Version:     "7.0.0",
	}, nil
}

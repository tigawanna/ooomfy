package smtp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/devstack/devstack/internal/config"
)

type Email struct {
	ID      string    `json:"id"`
	From    string    `json:"from"`
	To      []string  `json:"to"`
	Subject string    `json:"subject"`
	Date    time.Time `json:"date"`
	Body    string    `json:"body"`
	Raw     string    `json:"raw,omitempty"`
}

type Service struct {
	mu       sync.RWMutex
	cfg      *config.SMTPConfig
	persist  *config.SMTPPersistConfig
	running  bool
	listener net.Listener
	emails   map[string]*Email
	nextID   int
}

func New(cfg *config.SMTPConfig, persist *config.SMTPPersistConfig) *Service {
	return &Service{
		cfg:     cfg,
		persist: persist,
		running: false,
		emails:  make(map[string]*Email),
		nextID:  1,
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
		log.Printf("SMTP: port %d already in use, trying %d", port, port+1)
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
		log.Printf("SMTP: requested port unavailable, using port %d instead", port)
	}

	if err := os.MkdirAll(s.persist.Directory, 0755); err != nil {
		return fmt.Errorf("failed to create SMTP data directory: %w", err)
	}

	s.loadEmails()

	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	s.listener = listener
	s.running = true

	go s.acceptLoop()

	log.Printf("SMTP: started on %s (data dir: %s)", addr, s.persist.Directory)
	return nil
}

func (s *Service) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	if s.listener != nil {
		s.listener.Close()
	}

	s.running = false
	s.listener = nil

	log.Println("SMTP: stopped")
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

func (s *Service) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}
		go s.handleConnection(conn)
	}
}

func (s *Service) handleConnection(conn net.Conn) {
	defer conn.Close()

	s.mu.Lock()
	email := &Email{Date: time.Now()}
	s.nextID++
	email.ID = fmt.Sprintf("%d", s.nextID)
	s.emails[email.ID] = email
	s.mu.Unlock()

	var (
		state    int
		mailFrom string
		rcptTos  []string
		data     []byte
	)

	conn.Write([]byte("220 DevStack SMTP Server\r\n"))

	buffer := make([]byte, 4096)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if err != io.EOF {
				log.Printf("SMTP read error: %v", err)
			}
			return
		}

		line := string(buffer[:n])
		lines := strings.Split(line, "\n")

		for _, l := range lines {
			l = strings.TrimSuffix(l, "\r")
			cmd := strings.Fields(l)
			if len(cmd) == 0 {
				continue
			}

			verb := strings.ToUpper(cmd[0])
			args := ""
			if len(cmd) > 1 {
				args = strings.Join(cmd[1:], " ")
			}

			switch state {
			case 0:
				switch verb {
				case "HELO", "EHLO":
					conn.Write([]byte("250 Hello\r\n"))
				case "MAIL":
					mailFrom = extractEmail(strings.TrimPrefix(args, "FROM:"))
					state = 1
					conn.Write([]byte("250 OK\r\n"))
				case "QUIT":
					conn.Write([]byte("221 Bye\r\n"))
					return
				default:
					conn.Write([]byte("500 Unknown command\r\n"))
				}
			case 1:
				switch verb {
				case "MAIL":
					mailFrom = extractEmail(strings.TrimPrefix(args, "FROM:"))
					conn.Write([]byte("250 OK\r\n"))
				case "RCPT":
					rcptTos = append(rcptTos, extractEmail(strings.TrimPrefix(args, "TO:")))
					conn.Write([]byte("250 OK\r\n"))
				case "DATA":
					state = 2
					conn.Write([]byte("354 Start mail input\r\n"))
				case "QUIT":
					conn.Write([]byte("221 Bye\r\n"))
					return
				default:
					conn.Write([]byte("500 Unknown command\r\n"))
				}
			case 2:
				if l == "." {
					state = 1
					email.From = mailFrom
					email.To = rcptTos
					email.Raw = string(data)

					if msg, err := mail.ReadMessage(bytes.NewReader(data)); err == nil {
						email.Subject = msg.Header.Get("Subject")
						if body, err := io.ReadAll(msg.Body); err == nil {
							email.Body = string(body)
						}
					}

					if email.Subject == "" {
						email.Subject = "(no subject)"
					}

					s.saveEmail(email)
					s.mu.Lock()
					s.emails[email.ID] = email
					s.mu.Unlock()

					conn.Write([]byte("250 OK\r\n"))
					data = nil
				} else {
					data = append(data, []byte(l+"\r\n")...)
				}
			}
		}
	}
}

func extractEmail(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "<")
	if idx := strings.Index(s, ">"); idx >= 0 {
		return s[:idx]
	}
	return s
}

func (s *Service) loadEmails() {
	entries, err := os.ReadDir(s.persist.Directory)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(s.persist.Directory, entry.Name()))
		if err != nil {
			continue
		}

		var email Email
		if err := json.Unmarshal(data, &email); err != nil {
			continue
		}

		s.emails[email.ID] = &email

		var id int
		fmt.Sscanf(email.ID, "%d", &id)
		if id >= s.nextID {
			s.nextID = id + 1
		}
	}
}

func (s *Service) saveEmail(email *Email) {
	if !s.persist.Enabled {
		return
	}

	data, err := json.MarshalIndent(email, "", "  ")
	if err != nil {
		log.Printf("SMTP: failed to marshal email: %v", err)
		return
	}

	filename := filepath.Join(s.persist.Directory, email.ID+".json")
	tmpFile := filename + ".tmp"

	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		log.Printf("SMTP: failed to write email: %v", err)
		return
	}

	os.Rename(tmpFile, filename)
}

func (s *Service) ListEmails() []Email {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Email, 0, len(s.emails))
	for _, email := range s.emails {
		result = append(result, *email)
	}
	return result
}

func (s *Service) GetEmail(id string) *Email {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.emails[id]
}

func (s *Service) DeleteEmail(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.emails[id]; !ok {
		return fmt.Errorf("email not found")
	}

	filename := filepath.Join(s.persist.Directory, id+".json")
	os.Remove(filename)
	delete(s.emails, id)
	return nil
}

func (s *Service) ClearEmails() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id := range s.emails {
		filename := filepath.Join(s.persist.Directory, id+".json")
		os.Remove(filename)
	}

	s.emails = make(map[string]*Email)
	return nil
}

type Stats struct {
	NumEmails int `json:"numEmails"`
}

func (s *Service) GetStats() *Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return &Stats{NumEmails: len(s.emails)}
}

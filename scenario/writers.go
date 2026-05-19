package scenario

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// OutputMode selects how scenario logs are written.
type OutputMode string

const (
	OutputInterleaved OutputMode = "interleaved"
	OutputSplit       OutputMode = "split"
	OutputBoth        OutputMode = "both"
)

// WritersConfig configures scenario output destinations.
type WritersConfig struct {
	Mode      OutputMode
	Combined  io.Writer // interleaved or both
	OutputDir string    // split or both
}

// writerMux routes encoded lines to combined and/or per-service writers.
type writerMux struct {
	mode      OutputMode
	combined  io.Writer
	byService map[string]io.Writer
	closers   []func()
}

func newWriterMux(cfg WritersConfig, services []string) (*writerMux, error) {
	mode := cfg.Mode
	if mode == "" {
		mode = OutputInterleaved
	}
	m := &writerMux{mode: mode, byService: make(map[string]io.Writer)}
	switch mode {
	case OutputInterleaved:
		if cfg.Combined == nil {
			return nil, fmt.Errorf("scenario: combined output writer is required")
		}
		m.combined = cfg.Combined
	case OutputSplit:
		if cfg.OutputDir == "" {
			return nil, fmt.Errorf("scenario: output-dir is required for split mode")
		}
		if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
			return nil, err
		}
		for _, svc := range services {
			path := filepath.Join(cfg.OutputDir, sanitizeFilename(svc)+".jsonl")
			f, err := os.Create(path)
			if err != nil {
				m.closeAll()
				return nil, err
			}
			m.byService[svc] = f
			m.closers = append(m.closers, func() { _ = f.Close() })
		}
	case OutputBoth:
		if cfg.Combined == nil {
			return nil, fmt.Errorf("scenario: combined output writer is required for both mode")
		}
		if cfg.OutputDir == "" {
			return nil, fmt.Errorf("scenario: output-dir is required for both mode")
		}
		m.combined = cfg.Combined
		if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
			return nil, err
		}
		for _, svc := range services {
			path := filepath.Join(cfg.OutputDir, sanitizeFilename(svc)+".jsonl")
			f, err := os.Create(path)
			if err != nil {
				m.closeAll()
				return nil, err
			}
			m.byService[svc] = f
			m.closers = append(m.closers, func() { _ = f.Close() })
		}
	default:
		return nil, fmt.Errorf("scenario: unknown output mode %q", mode)
	}
	return m, nil
}

func (m *writerMux) write(service string, line []byte) error {
	switch m.mode {
	case OutputInterleaved:
		_, err := m.combined.Write(line)
		return err
	case OutputSplit:
		w := m.byService[service]
		if w == nil {
			return fmt.Errorf("scenario: no writer for service %q", service)
		}
		_, err := w.Write(line)
		return err
	case OutputBoth:
		if _, err := m.combined.Write(line); err != nil {
			return err
		}
		w := m.byService[service]
		if w == nil {
			return fmt.Errorf("scenario: no writer for service %q", service)
		}
		_, err := w.Write(line)
		return err
	}
	return nil
}

func (m *writerMux) closeAll() {
	for _, c := range m.closers {
		c()
	}
}

func sanitizeFilename(s string) string {
	s = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, s)
	if s == "" {
		return "service"
	}
	return s
}

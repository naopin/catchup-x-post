package history

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const entriesFile = "entries.jsonl"

// Entry は過去に生成（または投稿）したネタの記録
type Entry struct {
	URL       string    `json:"url"`
	Title     string    `json:"title"`
	Topic     string    `json:"topic"`
	CreatedAt time.Time `json:"created_at"`
}

type Store struct {
	dir     string
	entries []Entry
}

func NewStore(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	s := &Store{dir: dir}
	if err := s.load(); err != nil {
		return nil, err
	}
	s.migrateLegacyPostedURLs()
	return s, nil
}

func (s *Store) load() error {
	path := filepath.Join(s.dir, entriesFile)
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var e Entry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			continue
		}
		if e.URL != "" {
			s.entries = append(s.entries, e)
		}
	}
	return sc.Err()
}

func (s *Store) migrateLegacyPostedURLs() {
	legacy := filepath.Join("data", "posted_urls.txt")
	f, err := os.Open(legacy)
	if err != nil {
		return
	}
	defer f.Close()

	var added bool
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		url := strings.TrimSpace(sc.Text())
		if url == "" || s.hasURL(url) {
			continue
		}
		s.entries = append(s.entries, Entry{URL: url, CreatedAt: time.Now()})
		added = true
	}
	if added {
		_ = s.persist()
	}
}

func (s *Store) hasURL(url string) bool {
	for _, e := range s.entries {
		if e.URL == url {
			return true
		}
	}
	return false
}

func (s *Store) Count() int {
	return len(s.entries)
}

func (s *Store) Add(url, title, topic string) error {
	url = strings.TrimSpace(url)
	if url == "" {
		return nil
	}
	e := Entry{
		URL:       url,
		Title:     strings.TrimSpace(title),
		Topic:     strings.TrimSpace(topic),
		CreatedAt: time.Now(),
	}
	if s.IsSimilarEntry(e) {
		return nil
	}
	s.entries = append(s.entries, e)
	return s.persist()
}

func (s *Store) persist() error {
	path := filepath.Join(s.dir, entriesFile)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, e := range s.entries {
		if err := enc.Encode(e); err != nil {
			return err
		}
	}
	return nil
}

// IsSimilar は過去エントリと同一 URL または同様の記事かを判定する
func (s *Store) IsSimilar(url, title, topic string) bool {
	return s.IsSimilarEntry(Entry{
		URL:   strings.TrimSpace(url),
		Title: strings.TrimSpace(title),
		Topic: strings.TrimSpace(topic),
	})
}

func (s *Store) IsSimilarEntry(candidate Entry) bool {
	for _, past := range s.entries {
		if similarEntry(past, candidate) {
			return true
		}
	}
	return false
}

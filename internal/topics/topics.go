package topics

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const historyFile = "topics_history.txt"

type Store struct {
	path   string
	titles []string
}

func NewStore(logsDir string) (*Store, error) {
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, err
	}
	s := &Store{path: filepath.Join(logsDir, historyFile)}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) load() error {
	f, err := os.Open(s.path)
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
		title, _ := parseLine(line)
		if title != "" {
			s.titles = append(s.titles, title)
		}
	}
	return sc.Err()
}

func parseLine(line string) (title, url string) {
	prefixes := []string{"選んだトピック:", "選定トピック：", "選定トピック:"}
	for _, p := range prefixes {
		if strings.HasPrefix(line, p) {
			rest := strings.TrimSpace(strings.TrimPrefix(line, p))
			if idx := strings.Index(rest, "https://"); idx >= 0 {
				return strings.TrimSpace(rest[:idx]), strings.TrimSpace(rest[idx:])
			}
			return rest, ""
		}
	}
	if idx := strings.Index(line, "https://"); idx >= 0 {
		return strings.TrimSpace(line[:idx]), strings.TrimSpace(line[idx:])
	}
	return line, ""
}

func (s *Store) Count() int {
	return len(s.titles)
}

// IsDuplicate は topics_history に類似トピックがあるか
func (s *Store) IsDuplicate(title, topic string) bool {
	candidate := topicLabel(title, topic)
	if candidate == "" {
		return false
	}
	for _, past := range s.titles {
		if historyTopicSimilar(past, candidate) || historyTopicSimilar(past, title) || historyTopicSimilar(past, topic) {
			return true
		}
	}
	return false
}

func historyTopicSimilar(a, b string) bool {
	a = normalizeTopicKey(a)
	b = normalizeTopicKey(b)
	if a == "" || b == "" {
		return false
	}
	if a == b {
		return true
	}
	if strings.Contains(a, b) || strings.Contains(b, a) {
		return true
	}
	return false
}

func normalizeTopicKey(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func topicLabel(title, topic string) string {
	title = strings.TrimSpace(title)
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return title
	}
	if title == "" || title == topic {
		return topic
	}
	if strings.Contains(title, topic) {
		return title
	}
	return topic + " (" + title + ")"
}

// Append は選定トピックを1行追記する
func (s *Store) Append(title, topic, url string) error {
	line := FormatLine(title, topic, url)
	if line == "" {
		return nil
	}
	label, _ := parseLine(line)
	for _, past := range s.titles {
		if historyTopicSimilar(past, label) {
			return nil
		}
	}
	s.titles = append(s.titles, label)

	f, err := os.OpenFile(s.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintln(f, line)
	return err
}

// FormatLine は topics_history.txt 用の1行を返す
func FormatLine(title, topic, url string) string {
	url = strings.TrimSpace(url)
	label := topicLabel(title, topic)
	if label == "" {
		return ""
	}
	if url == "" {
		return "選んだトピック: " + label
	}
	return "選んだトピック: " + label + " " + url
}

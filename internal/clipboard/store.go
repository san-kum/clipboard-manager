package clipboard

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type ClipboardStore struct {
	mu          sync.Mutex
	maxEntries  int
	storagePath string
	log         *logrus.Logger
	entries     []*ClipboardEntry
}

func NewClipboardStore(storagePath string, maxEntries int, logger *logrus.Logger) *ClipboardStore {
	return &ClipboardStore{
		storagePath: storagePath,
		maxEntries:  maxEntries,
		log:         logger,
		entries:     []*ClipboardEntry{},
	}
}

func (cs *ClipboardStore) Add(entry *ClipboardEntry) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if len(cs.entries) >= cs.maxEntries {
		cs.entries = cs.entries[1:]
	}
	cs.entries = append(cs.entries, entry)

	return cs.persist()
}

func (cs *ClipboardStore) List() []*ClipboardEntry {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	return cs.entries
}

func (cs *ClipboardStore) persist() error {

	filename := filepath.Join(cs.storagePath, "clipboard_history.json")

	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		return errors.Wrap(err, "failed to create a storage directory")
	}

	data, err := json.MarshalIndent(cs.entries, "", " ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal entries")
	}
	return os.WriteFile(filename, data, 0644)
}

func (cs *ClipboardStore) Load() error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	filename := filepath.Join(cs.storagePath, "clipboard_history.json")

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return errors.Wrap(err, "failed to read clipboard history")
	}

	var entries []*ClipboardEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return errors.Wrap(err, "failed to parse clipboard history")
	}

	cs.entries = entries
	return nil
}

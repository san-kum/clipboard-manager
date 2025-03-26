package clipboard

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/atotto/clipboard"
	"github.com/sirupsen/logrus"
)

type ClipboardEntry struct {
	ID        string
	Content   string
	Timestamp time.Time
	Type      string
}

type ClipboardManager struct {
	currentContent string
	store          *ClipboardStore
	log            *logrus.Logger
	cancel         context.CancelFunc
	mu             sync.Mutex
}

func NewClipboardManager(store *ClipboardStore, logger *logrus.Logger) *ClipboardManager {
	return &ClipboardManager{
		store: store,
		log:   logger,
	}
}

func (cm *ClipboardManager) Start(ctx context.Context) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	ctx, cancel := context.WithCancel(ctx)
	cm.cancel = cancel

	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				cm.log.Info("clipboard listener stopped")
				return
			case <-ticker.C:
				cm.checkClipboardChange()

			}
		}
	}()

	cm.log.Info("Clipboard listener started")
	return nil

}

func (cm *ClipboardManager) Stop() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.cancel != nil {
		cm.cancel()
	}
}

func (cm *ClipboardManager) checkClipboardChange() {
	content, err := clipboard.ReadAll()
	if err != nil {
		cm.log.WithError(err).Error("Failed to read clipboard")
		return
	}

	if content == "" {
		return
	}

	if content == cm.currentContent {
		return
	}

	cm.currentContent = content

	contentType := cm.detectConentType(content)

	entry := &ClipboardEntry{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Content:   content,
		Timestamp: time.Now(),
		Type:      contentType,
	}

	if err := cm.store.Add(entry); err != nil {
		cm.log.WithError(err).Error("Failed to store clipboard entry")
	}

	cm.log.WithFields(logrus.Fields{
		"type":      contentType,
		"length":    len(content),
		"timestamp": entry.Timestamp,
	}).Info("New clipboard entry captured")
}

func (cm *ClipboardManager) detectConentType(content string) string {
	switch {
	case len(content) > 1000:
		return "longtext"
	case len(content) < 10:
		return "shorttext"
	default:
		return "text"
	}
}

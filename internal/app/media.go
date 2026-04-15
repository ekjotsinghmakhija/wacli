// internal/app/media.go
package app

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/events"
)

type MediaDownloader struct {
	client *whatsmeow.Client
	dir    string
	jobs   chan *events.Message
}

func NewMediaDownloader(client *whatsmeow.Client, storeDir string) *MediaDownloader {
	md := &MediaDownloader{
		client: client,
		dir:    filepath.Join(storeDir, "media"),
		jobs:   make(chan *events.Message, 100),
	}
	_ = os.MkdirAll(md.dir, 0700)
	return md
}

func (m *MediaDownloader) Start(ctx context.Context, workers int) {
	for i := 0; i < workers; i++ {
		go m.worker(ctx)
	}
}

func (m *MediaDownloader) Queue(msg *events.Message) {
	select {
	case m.jobs <- msg:
	default:
		log.Println("media download queue full, dropping job")
	}
}

func (m *MediaDownloader) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-m.jobs:
			m.download(msg)
		}
	}
}

func (m *MediaDownloader) download(evt *events.Message) {
	msg := evt.Message
	var data []byte
	var err error
	var ext string

	if doc := msg.GetDocumentMessage(); doc != nil {
		data, err = m.client.Download(doc)
		ext = ".doc"
	} else if img := msg.GetImageMessage(); img != nil {
		data, err = m.client.Download(img)
		ext = ".jpg"
	} else if vid := msg.GetVideoMessage(); vid != nil {
		data, err = m.client.Download(vid)
		ext = ".mp4"
	} else if audio := msg.GetAudioMessage(); audio != nil {
		data, err = m.client.Download(audio)
		ext = ".ogg"
	}

	if err != nil {
		log.Printf("media download failed for %s: %v", evt.Info.ID, err)
		return
	}
	if len(data) == 0 {
		return
	}

	path := filepath.Join(m.dir, evt.Info.ID+ext)
	_ = os.WriteFile(path, data, 0600)
}

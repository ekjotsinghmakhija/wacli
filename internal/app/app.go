// internal/app/app.go
package app

import (
	"context"

	"github.com/ekjotsinghmakhija/wacli/internal/store"
	"github.com/ekjotsinghmakhija/wacli/internal/wa"
)

type App struct {
	db *store.DB
	wa wa.WAClient
}

func New(db *store.DB, client wa.WAClient) *App {
	return &App{
		db: db,
		wa: client,
	}
}

func (a *App) StartSync(ctx context.Context) error {
	wa.BindRouter(a.wa, a)
	if !a.wa.IsConnected() {
		if err := a.wa.Connect(); err != nil {
			return err
		}
	}
	<-ctx.Done()
	a.wa.Disconnect()
	return nil
}

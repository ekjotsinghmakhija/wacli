package main

import (
	"fmt"

	"github.com/ekjotsinghmakhija/wacli/internal/app"
	"github.com/ekjotsinghmakhija/wacli/internal/config"
	"github.com/ekjotsinghmakhija/wacli/internal/lock"
	"github.com/ekjotsinghmakhija/wacli/internal/store"
	"github.com/ekjotsinghmakhija/wacli/internal/wa"
	"github.com/spf13/cobra"
)

func newSyncCmd(f *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Start the background sync daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			storeDir := f.storeDir
			if storeDir == "" {
				storeDir = config.DefaultStoreDir()
			}

			l, err := lock.Acquire(storeDir)
			if err != nil {
				return err
			}
			defer l.Release()

			db, err := store.Open(storeDir + "/wacli.db")
			if err != nil {
				return wrapErr(err, "open db")
			}
			defer db.Close()

			client, err := wa.New(storeDir, false)
			if err != nil {
				return wrapErr(err, "init client")
			}

			if !client.IsLoggedIn() {
				return fmt.Errorf("not logged in. Run 'wacli login' first")
			}

			a := app.New(db, client)

			ctx, cancel := signalContext()
			defer cancel()

			fmt.Println("Starting sync daemon. Press Ctrl+C to stop.")
			return a.StartSync(ctx)
		},
	}
}

// cmd/wacli/doctor.go
package main

import (
	"fmt"
	"os"

	"github.com/ekjotsinghmakhija/wacli/internal/config"
	"github.com/ekjotsinghmakhija/wacli/internal/lock"
	"github.com/ekjotsinghmakhija/wacli/internal/store"
	"github.com/ekjotsinghmakhija/wacli/internal/wa"
	"github.com/spf13/cobra"
)

func newDoctorCmd(f *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check system health, database integrity, and locks",
		RunE: func(cmd *cobra.Command, args []string) error {
			storeDir := f.storeDir
			if storeDir == "" {
				storeDir = config.DefaultStoreDir()
			}

			fmt.Println("Diagnosing WaCLI Configuration...")
			fmt.Printf("- Store Directory: %s\n", storeDir)

			if _, err := os.Stat(storeDir); os.IsNotExist(err) {
				fmt.Println("- Status: [WARN] Store directory does not exist yet.")
			} else {
				fmt.Println("- Status: [OK] Store directory found.")
			}

			l, err := lock.Acquire(storeDir)
			if err != nil {
				fmt.Printf("- Lock Status: [LOCKED] Error: %v\n  (A sync daemon might be running in another terminal)\n", err)
			} else {
				fmt.Println("- Lock Status: [FREE] Successfully acquired system lock.")
				l.Release()
			}

			db, err := store.Open(storeDir + "/wacli.db")
			if err != nil {
				fmt.Printf("- DB Status: [ERROR] %v\n", err)
			} else {
				if db.HasFTS() {
					fmt.Println("- DB Status: [OK] Connected. SQLite FTS5 extension is ACTIVE.")
				} else {
					fmt.Println("- DB Status: [WARN] Connected, but FTS5 extension is missing.")
				}
				db.Close()
			}

			client, err := wa.New(storeDir, false)
			if err != nil {
				fmt.Printf("- WA Status: [ERROR] %v\n", err)
			} else {
				if client.IsLoggedIn() {
					fmt.Println("- WA Status: [OK] Session token found. Authenticated.")
				} else {
					fmt.Println("- WA Status: [WARN] No active session. Run 'wacli login'.")
				}
				client.Disconnect()
			}

			return nil
		},
	}
}

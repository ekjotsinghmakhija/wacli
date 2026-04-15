// cmd/wacli/chats.go
package main

import (
	"time"

	"github.com/ekjotsinghmakhija/wacli/internal/config"
	"github.com/ekjotsinghmakhija/wacli/internal/out"
	"github.com/ekjotsinghmakhija/wacli/internal/store"
	"github.com/spf13/cobra"
)

func newChatsCmd(f *rootFlags) *cobra.Command {
	var limit int
	var query string

	cmd := &cobra.Command{
		Use:   "chats",
		Short: "List recent chats",
		RunE: func(cmd *cobra.Command, args []string) error {
			storeDir := f.storeDir
			if storeDir == "" {
				storeDir = config.DefaultStoreDir()
			}

			db, err := store.Open(storeDir + "/wacli.db")
			if err != nil {
				return wrapErr(err, "open db")
			}
			defer db.Close()

			chats, err := db.ListChats(query, limit)
			if err != nil {
				return wrapErr(err, "list chats")
			}

			p := out.New(f.asJSON)
			p.Print(chats, []string{"JID", "NAME", "KIND", "LAST MESSAGE"}, func(data interface{}) [][]string {
				list := data.([]store.Chat)
				var rows [][]string
				for _, c := range list {
					ts := ""
					if !c.LastMessageTS.IsZero() {
						ts = c.LastMessageTS.Local().Format(time.RFC3339)
					}
					rows = append(rows, []string{c.JID, c.Name, c.Kind, ts})
				}
				return rows
			})
			return nil
		},
	}
	cmd.Flags().IntVarP(&limit, "limit", "l", 50, "Limit number of results")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Filter chats by name or JID")
	return cmd
}

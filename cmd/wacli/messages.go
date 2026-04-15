// cmd/wacli/messages.go
package main

import (
	"time"

	"github.com/ekjotsinghmakhija/wacli/internal/config"
	"github.com/ekjotsinghmakhija/wacli/internal/out"
	"github.com/ekjotsinghmakhija/wacli/internal/store"
	"github.com/spf13/cobra"
)

func newMessagesCmd(f *rootFlags) *cobra.Command {
	var limit int
	var chatJID string

	cmd := &cobra.Command{
		Use:   "messages [query]",
		Short: "Search or list messages",
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

			query := ""
			if len(args) > 0 {
				query = args[0]
			}

			msgs, err := db.SearchMessages(store.SearchMessagesParams{
				Query:   query,
				ChatJID: chatJID,
				Limit:   limit,
			})
			if err != nil {
				return wrapErr(err, "search messages")
			}

			p := out.New(f.asJSON)
			p.Print(msgs, []string{"TIME", "CHAT", "SENDER", "MESSAGE"}, func(data interface{}) [][]string {
				list := data.([]store.Message)
				var rows [][]string
				for _, m := range list {
					ts := m.Timestamp.Local().Format("2006-01-02 15:04:05")
					chat := m.ChatName
					if chat == "" {
						chat = m.ChatJID
					}
					sender := "Me"
					if !m.FromMe {
						sender = m.SenderJID
					}
					text := m.Snippet
					if text == "" {
						text = m.DisplayText
					}
					rows = append(rows, []string{ts, chat, sender, text})
				}
				return rows
			})
			return nil
		},
	}
	cmd.Flags().IntVarP(&limit, "limit", "l", 50, "Limit number of results")
	cmd.Flags().StringVarP(&chatJID, "chat", "c", "", "Filter by specific Chat JID")
	return cmd
}

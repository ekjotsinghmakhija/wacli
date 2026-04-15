// cmd/wacli/send.go
package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/ekjotsinghmakhija/wacli/internal/config"
	"github.com/ekjotsinghmakhija/wacli/internal/out"
	"github.com/ekjotsinghmakhija/wacli/internal/wa"
	"github.com/spf13/cobra"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

func newSendCmd(f *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send [jid] [message]",
		Short: "Send a text message",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			storeDir := f.storeDir
			if storeDir == "" {
				storeDir = config.DefaultStoreDir()
			}

			client, err := wa.New(storeDir, false)
			if err != nil {
				return wrapErr(err, "init client")
			}
			defer client.Disconnect()

			if !client.IsLoggedIn() {
				return fmt.Errorf("not logged in. Run 'wacli login' first")
			}

			if err := client.Connect(); err != nil {
				return wrapErr(err, "connect")
			}

			jid := wa.ParseJID(args[0])
			text := strings.Join(args[1:], " ")

			ctx, cancel := context.WithTimeout(cmd.Context(), f.timeout)
			defer cancel()

			msg := &waProto.Message{
				Conversation: proto.String(text),
			}

			resp, err := client.Client().SendMessage(ctx, jid, msg)
			if err != nil {
				return wrapErr(err, "send message")
			}

			p := out.New(f.asJSON)
			if f.asJSON {
				p.Print(map[string]interface{}{"status": "success", "msg_id": resp.ID, "timestamp": resp.Timestamp.Unix()}, nil, nil)
			} else {
				fmt.Printf("Message sent successfully. ID: %s\n", resp.ID)
			}
			return nil
		},
	}
	return cmd
}

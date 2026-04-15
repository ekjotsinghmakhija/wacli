// cmd/wacli/send_file.go
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ekjotsinghmakhija/wacli/internal/config"
	"github.com/ekjotsinghmakhija/wacli/internal/out"
	"github.com/ekjotsinghmakhija/wacli/internal/wa"
	"github.com/spf13/cobra"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

func newSendFileCmd(f *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "send-file [jid] [filepath]",
		Short: "Send a media file or document",
		Args:  cobra.ExactArgs(2),
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
			filePath := args[1]

			data, err := os.ReadFile(filePath)
			if err != nil {
				return wrapErr(err, "read file")
			}

			mimeType := http.DetectContentType(data)
			if ext := filepath.Ext(filePath); ext == ".json" {
				mimeType = "application/json"
			} else if ext == ".pdf" {
				mimeType = "application/pdf"
			} else if ext == ".txt" || ext == ".md" {
				mimeType = "text/plain"
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), f.timeout)
			defer cancel()

			var waMediaType whatsmeow.MediaType
			msg := &waProto.Message{}

			if strings.HasPrefix(mimeType, "image/") {
				waMediaType = whatsmeow.MediaImage
				resp, uploadErr := client.Client().Upload(ctx, data, waMediaType)
				if uploadErr != nil {
					return wrapErr(uploadErr, "upload image")
				}
				msg.ImageMessage = &waProto.ImageMessage{
					Url:           proto.String(resp.URL),
					DirectPath:    proto.String(resp.DirectPath),
					MediaKey:      resp.MediaKey,
					Mimetype:      proto.String(mimeType),
					FileEncSha256: resp.FileEncSHA256,
					FileSha256:    resp.FileSHA256,
					FileLength:    proto.Uint64(uint64(len(data))),
				}
			} else if strings.HasPrefix(mimeType, "video/") {
				waMediaType = whatsmeow.MediaVideo
				resp, uploadErr := client.Client().Upload(ctx, data, waMediaType)
				if uploadErr != nil {
					return wrapErr(uploadErr, "upload video")
				}
				msg.VideoMessage = &waProto.VideoMessage{
					Url:           proto.String(resp.URL),
					DirectPath:    proto.String(resp.DirectPath),
					MediaKey:      resp.MediaKey,
					Mimetype:      proto.String(mimeType),
					FileEncSha256: resp.FileEncSHA256,
					FileSha256:    resp.FileSHA256,
					FileLength:    proto.Uint64(uint64(len(data))),
				}
			} else {
				waMediaType = whatsmeow.MediaDocument
				resp, uploadErr := client.Client().Upload(ctx, data, waMediaType)
				if uploadErr != nil {
					return wrapErr(uploadErr, "upload document")
				}
				fileName := filepath.Base(filePath)
				msg.DocumentMessage = &waProto.DocumentMessage{
					Url:           proto.String(resp.URL),
					DirectPath:    proto.String(resp.DirectPath),
					MediaKey:      resp.MediaKey,
					Mimetype:      proto.String(mimeType),
					FileEncSha256: resp.FileEncSHA256,
					FileSha256:    resp.FileSHA256,
					FileLength:    proto.Uint64(uint64(len(data))),
					Title:         proto.String(fileName),
					FileName:      proto.String(fileName),
				}
			}

			resp, err := client.Client().SendMessage(ctx, jid, msg)
			if err != nil {
				return wrapErr(err, "send file message")
			}

			p := out.New(f.asJSON)
			if f.asJSON {
				p.Print(map[string]interface{}{"status": "success", "msg_id": resp.ID, "mime_type": mimeType}, nil, nil)
			} else {
				fmt.Printf("File sent successfully. ID: %s (Detected MIME: %s)\n", resp.ID, mimeType)
			}
			return nil
		},
	}
}

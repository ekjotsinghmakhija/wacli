package main

import (
	"fmt"
	"os"

	"github.com/ekjotsinghmakhija/wacli/internal/config"
	"github.com/ekjotsinghmakhija/wacli/internal/wa"
	"github.com/spf13/cobra"
)

func newAuthCmd(f *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Link a new WhatsApp device via QR code",
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

			if client.IsLoggedIn() {
				fmt.Println("Already logged in.")
				return nil
			}

			ctx, cancel := withTimeout(cmd.Context(), f)
			defer cancel()

			qrChan, err := client.GetQRChannel(ctx)
			if err != nil {
				return wrapErr(err, "get qr channel")
			}

			fmt.Println("Scan the QR code below using WhatsApp on your phone:")
			for evt := range qrChan {
				if evt.Event == "code" {
					fmt.Println(evt.Code) // In a real TUI, use qrcode-terminal or similar here
					fmt.Println("--- Waiting for scan ---")
				} else {
					fmt.Printf("Login event: %s\n", evt.Event)
				}
			}
			return nil
		},
	}
}

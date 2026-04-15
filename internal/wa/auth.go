package wa

import (
	"context"
	"fmt"

	"go.mau.fi/whatsmeow"
)

// GetQRChannel initiates the pairing process and returns a channel of QR codes.
func (c *clientImpl) GetQRChannel(ctx context.Context) (<-chan whatsmeow.QRChannelItem, error) {
	if c.IsLoggedIn() {
		return nil, fmt.Errorf("already logged in")
	}

	// Request QR channel. Whatsmeow automatically rotates the QR code.
	qrChan, err := c.cli.GetQRChannel(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get QR channel: %w", err)
	}

	if err := c.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect for pairing: %w", err)
	}

	return qrChan, nil
}

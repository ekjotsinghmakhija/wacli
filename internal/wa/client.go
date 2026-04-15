// internal/wa/client.go
package wa

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// WAClient defines the strict boundary for the WhatsApp network layer.
type WAClient interface {
	Connect() error
	Disconnect()
	IsConnected() bool
	IsLoggedIn() bool
	GetQRChannel(ctx context.Context) (<-chan whatsmeow.QRChannelItem, error)
	Logout() error
	AddEventHandler(handler whatsmeow.EventHandler) uint32
	RemoveEventHandler(id uint32)
	Client() *whatsmeow.Client
}

type clientImpl struct {
	cli         *whatsmeow.Client
	sessionPath string
}

// New initializes an isolated whatsmeow client using a separate session database.
func New(storeDir string, debug bool) (WAClient, error) {
	logLevel := "INFO"
	if debug {
		logLevel = "DEBUG"
	}
	dbLog := waLog.Stdout("Database", logLevel, true)

	// CRITICAL: Session DB is kept strictly separate from the application DB (wacli.db)
	sessionDBPath := filepath.Join(storeDir, "session.db")
	dsn := fmt.Sprintf("file:%s?_foreign_keys=on", sessionDBPath)

	// Inject context.Background() for newer versions of sqlstore
	container, err := sqlstore.New(context.Background(), "sqlite3", dsn, dbLog)
	if err != nil {
		return nil, fmt.Errorf("failed to init session store: %w", err)
	}

	// Inject context.Background() for GetFirstDevice
	deviceStore, err := container.GetFirstDevice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	clientLog := waLog.Stdout("Client", logLevel, true)
	cli := whatsmeow.NewClient(deviceStore, clientLog)

	return &clientImpl{
		cli:         cli,
		sessionPath: sessionDBPath,
	}, nil
}

func (c *clientImpl) Connect() error {
	return c.cli.Connect()
}

func (c *clientImpl) Disconnect() {
	c.cli.Disconnect()
}

func (c *clientImpl) IsConnected() bool {
	return c.cli.IsConnected()
}

func (c *clientImpl) IsLoggedIn() bool {
	return c.cli.Store.ID != nil
}

func (c *clientImpl) Logout() error {
	// Inject context.Background() for Logout
	err := c.cli.Logout(context.Background())
	if err != nil {
		// If logout fails, nuke the local session database to force clean state
		_ = os.Remove(c.sessionPath)
		return err
	}
	return nil
}

func (c *clientImpl) AddEventHandler(handler whatsmeow.EventHandler) uint32 {
	return c.cli.AddEventHandler(handler)
}

func (c *clientImpl) RemoveEventHandler(id uint32) {
	c.cli.RemoveEventHandler(id)
}

func (c *clientImpl) Client() *whatsmeow.Client {
	return c.cli
}

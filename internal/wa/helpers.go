// internal/wa/helpers.go
package wa

import (
	"time"

	"go.mau.fi/whatsmeow/types"
)

func ParseJID(jid string) types.JID {
	j, _ := types.ParseJID(jid)
	return j
}

func ConvertTime(ts uint64) time.Time {
	return time.Unix(int64(ts), 0)
}

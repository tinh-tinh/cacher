package sqlite3

import (
	"time"
)

func ParseTimestap(ttl time.Duration) time.Time {
	expiresAt := time.Now().UTC().Add(time.Duration(ttl))
	return expiresAt
}

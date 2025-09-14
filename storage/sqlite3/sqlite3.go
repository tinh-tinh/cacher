package sqlite3

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/tinh-tinh/cacher/v2"
)

type Sqlite struct {
	db  *sql.DB
	ttl time.Duration
}

type Options struct {
	Addr string
	Ttl  time.Duration
}

const CreateTable = `
CREATE TABLE IF NOT EXISTS cache (
    key TEXT PRIMARY KEY,
    value TEXT,
    expires_at DATETIME NOT NULL
);
`

func New(opt Options) cacher.Store {
	db, err := sql.Open("sqlite3", opt.Addr)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	if _, err := db.Exec(CreateTable); err != nil {
		fmt.Println(err)
		return nil
	}
	sqlite := &Sqlite{
		db:  db,
		ttl: opt.Ttl,
	}
	go sqlite.gc(1 * time.Second)
	return sqlite
}

func (s *Sqlite) SetOptions(option cacher.StoreOptions) {
	if option.Ttl > 0 {
		s.ttl = option.Ttl
	}
}

func (s *Sqlite) Name() string {
	return cacher.SQLITE3
}

func (s *Sqlite) Set(ctx context.Context, key string, val []byte, opts ...cacher.StoreOptions) error {
	var ttl time.Duration
	if len(opts) > 0 {
		ttl = opts[0].Ttl
	} else {
		ttl = s.ttl
	}
	_, err := s.db.ExecContext(ctx, "INSERT INTO cache (key, value, expires_at) VALUES (?, ?, ?) ", key, string(val), ParseTimestap(ttl))
	if err != nil {
		return err
	}

	return nil
}

func (s *Sqlite) Get(ctx context.Context, key string) ([]byte, error) {
	// Handler
	var val string
	err := s.db.QueryRowContext(ctx, "SELECT value FROM cache WHERE key = ? AND expires_at > DATETIME('now')", key).Scan(&val)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return []byte(val), nil
}

func (s *Sqlite) Delete(ctx context.Context, key string) error {
	// Handler
	_, err := s.db.ExecContext(ctx, "DELETE FROM cache WHERE key = ?", key)
	if err != nil {
		return err
	}

	return nil
}

func (s *Sqlite) Clear(ctx context.Context) error {
	// Handler
	_, err := s.db.ExecContext(ctx, "DELETE FROM cache")
	if err != nil {
		return err
	}
	return nil
}

func (s *Sqlite) gc(sleep time.Duration) {
	ticker := time.NewTimer(sleep)
	defer ticker.Stop()
	for range ticker.C {
		s.db.Exec("DELETE FROM cache WHERE expires_at < DATETIME('now')")
	}
}

func (s *Sqlite) GetConnect() interface{} {
	return s.db
}

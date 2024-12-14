package sqlite3

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/tinh-tinh/cacher"
)

type Sqlite[M any] struct {
	db          *sql.DB
	ttl         time.Duration
	hooks       []cacher.Hook
	CompressAlg cacher.CompressAlg
}

type Options struct {
	Addr        string
	Ttl         time.Duration
	CompressAlg cacher.CompressAlg
	Hooks       []cacher.Hook
}

const CreateTable = `
CREATE TABLE IF NOT EXISTS cache (
    key TEXT PRIMARY KEY,
    value TEXT,
    expires_at DATETIME NOT NULL
);
`

func New[M any](opt Options) cacher.Store[M] {
	db, err := sql.Open("sqlite3", opt.Addr)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	if _, err := db.Exec(CreateTable); err != nil {
		fmt.Println(err)
		return nil
	}
	sqlite := &Sqlite[M]{
		db:          db,
		ttl:         opt.Ttl,
		hooks:       opt.Hooks,
		CompressAlg: opt.CompressAlg,
	}
	go sqlite.gc(1 * time.Second)
	return sqlite
}

func (s *Sqlite[M]) SetOptions(option cacher.StoreOptions) {
	if option.CompressAlg != "" && cacher.IsValidAlg(option.CompressAlg) {
		s.CompressAlg = option.CompressAlg
	}

	if option.Ttl > 0 {
		s.ttl = option.Ttl
	}

	if option.Hooks != nil {
		s.hooks = option.Hooks
	}
}

func (s *Sqlite[M]) Set(ctx context.Context, key string, val M, opts ...cacher.StoreOptions) error {
	cacher.HandlerBeforeSet(s, key, val)

	var value interface{}
	valStr, err := json.Marshal(&val)
	if err != nil {
		return err
	}
	value = string(valStr)
	// Handler
	if s.CompressAlg != "" {
		b, err := cacher.Compress(val, s.CompressAlg)
		if err != nil {
			return err
		}
		value = b
	}

	var ttl time.Duration
	if len(opts) > 0 {
		ttl = opts[0].Ttl
	} else {
		ttl = s.ttl
	}
	_, err = s.db.ExecContext(ctx, "INSERT INTO cache (key, value, expires_at) VALUES (?, ?, ?) ", key, value, ParseTimestap(ttl))
	if err != nil {
		return err
	}

	cacher.HandlerAfterSet(s, key, val)
	return nil
}

func (s *Sqlite[M]) Get(ctx context.Context, key string) (M, error) {
	cacher.HandlerBeforeGet(s, key)

	// Handler
	var schema M
	var val string
	err := s.db.QueryRowContext(ctx, "SELECT value FROM cache WHERE key = ? AND expires_at > DATETIME('now')", key).Scan(&val)
	if err != nil {
		if err == sql.ErrNoRows {
			return *new(M), nil
		}
		return *new(M), err
	}
	err = json.Unmarshal([]byte(val), &schema)
	if err != nil {
		if s.CompressAlg != "" {
			schema, err = cacher.Decompress[M]([]byte(val), s.CompressAlg)
			if err != nil {
				return *new(M), err
			}
		} else {
			return *new(M), err
		}
	}

	cacher.HandlerAfterGet(s, key, schema)
	return schema, nil
}

func (s *Sqlite[M]) Delete(ctx context.Context, key string) error {
	cacher.HandlerBeforeDelete(s, key)
	// Handler
	s.db.ExecContext(ctx, "DELETE FROM cache WHERE key = ?", key)

	cacher.HandlerAfterDelete(s, key)
	return nil
}

func (s *Sqlite[M]) MGet(ctx context.Context, keys ...string) ([]M, error) {
	var output []M
	for _, key := range keys {
		schema, err := s.Get(ctx, key)
		if err != nil {
			return nil, err
		}
		output = append(output, schema)
	}
	return output, nil
}

func (s *Sqlite[M]) MSet(ctx context.Context, data ...cacher.Params[M]) error {
	for _, d := range data {
		err := s.Set(ctx, d.Key, d.Val, d.Options)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Sqlite[M]) Clear(ctx context.Context) error {
	// Handler
	s.db.ExecContext(ctx, "DELETE FROM cache")
	return nil
}

func (s *Sqlite[M]) GetHooks() []cacher.Hook {
	return s.hooks
}

func (s *Sqlite[M]) gc(sleep time.Duration) {
	ticker := time.NewTimer(sleep)
	defer ticker.Stop()
	for range ticker.C {
		s.db.Exec("DELETE FROM cache WHERE expires_at < DATETIME('now')")
	}
}

func (s *Sqlite[M]) GetConnect() interface{} {
	return s.db
}

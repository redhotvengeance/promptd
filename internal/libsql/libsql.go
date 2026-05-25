package libsql

import (
	"database/sql"

	queries "github.com/redhotvengeance/promptd/internal/libsql/data/sql"
	"github.com/redhotvengeance/promptd/internal/promptd"
	_ "github.com/tursodatabase/go-libsql"
)

type Datastore struct {
	DB      *sql.DB
	Queries *queries.Queries

	threadService promptd.ThreadService
}

func NewDatastore() *Datastore {
	return &Datastore{}
}

func (d *Datastore) Open() error {
	libsql, err := sql.Open("libsql", "file:./internal/libsql/data/local.db")

	d.DB = libsql
	d.Queries = queries.New(libsql)

	d.threadService = NewThreadService(d.Queries)

	return err
}

func (d *Datastore) Threads() promptd.ThreadService {
	return d.threadService
}

func (d *Datastore) Close() error {
	return d.DB.Close()
}

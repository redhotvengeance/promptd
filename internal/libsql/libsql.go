package libsql

import (
	"database/sql"

	queries "github.com/redhotvengeance/promptd/internal/libsql/data/sql"
	_ "github.com/tursodatabase/go-libsql"
)

type Datastore struct {
	DB      *sql.DB
	Queries *queries.Queries
}

func NewDatastore() *Datastore {
	return &Datastore{}
}

func (d *Datastore) Open() error {
	libsql, err := sql.Open("libsql", "file:./internal/libsql/data/local.db")

	d.DB = libsql
	d.Queries = queries.New(libsql)

	return err
}

func (d *Datastore) Close() error {
	return d.DB.Close()
}

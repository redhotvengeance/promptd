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

	messageStore promptd.MessageStore
	threadStore promptd.ThreadStore
	workspaceStore promptd.WorkspaceStore
}

func NewDatastore() *Datastore {
	return &Datastore{}
}

func (d *Datastore) Open() error {
	libsql, err := sql.Open("libsql", "file:./internal/libsql/data/local.db")

	d.DB = libsql
	d.Queries = queries.New(libsql)

	d.messageStore = NewMessageStore(d.Queries)
	d.threadStore = NewThreadStore(d.Queries)
	d.workspaceStore = NewWorkspaceStore(d.DB, d.Queries)

	return err
}

func (d *Datastore) Messages() promptd.MessageStore {
	return d.messageStore
}

func (d *Datastore) Threads() promptd.ThreadStore {
	return d.threadStore
}

func (d *Datastore) Workspaces() promptd.WorkspaceStore {
	return d.workspaceStore
}

func (d *Datastore) Close() error {
	return d.DB.Close()
}

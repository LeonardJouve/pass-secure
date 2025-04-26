package database

import (
	"context"
	"embed"
	"errors"
	"os"
	"path"

	"github.com/LeonardJouve/pass-secure/database/queries"
	"github.com/LeonardJouve/pass-secure/status"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

type Database struct {
	conn *pgx.Conn
	qry  *queries.Queries
	ctx  context.Context
}

var db *Database

func New(connectionURL string) (*Database, error) {
	oldDb, err := GetInstance()
	if err == nil {
		oldDb.Close()
	}

	db = &Database{
		ctx: context.Background(),
	}

	db.conn, err = pgx.Connect(db.ctx, connectionURL)
	if err != nil {
		return &Database{}, err
	}

	db.qry = queries.New(db.conn)

	return db, nil
}

func GetInstance() (*Database, error) {
	if db == nil || db.conn.IsClosed() {
		return nil, errors.New("no database connection")
	}

	return db, nil
}

func BeginTransaction(c *fiber.Ctx) (*queries.Queries, *context.Context, func(), bool) {
	db, err := GetInstance()
	if err != nil {
		status.InternalServerError(c, nil)
		return nil, nil, func() {}, false
	}

	return db.BeginTransaction(c)
}

func (d *Database) Close() {
	d.conn.Close(d.ctx)
}

func (d *Database) Migrate(migrations embed.FS, migrationsFolder string) error {
	if d.conn.IsClosed() {
		return errors.New("database connection closed")
	}

	ctx := context.Background()

	executable, err := os.Executable()
	if err != nil {
		return err
	}

	migrationsPath := path.Join(path.Dir(executable), migrationsFolder)

	migrationEntries, err := migrations.ReadDir(migrationsPath)
	if err != nil {
		return err
	}

	for _, migrationEntry := range migrationEntries {
		if migrationEntry.IsDir() {
			continue
		}
		content, err := migrations.ReadFile(path.Join(migrationsPath, migrationEntry.Name()))
		if err != nil {
			return err
		}

		_, err = d.conn.Exec(ctx, string(content))
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Database) BeginTransaction(c *fiber.Ctx) (*queries.Queries, *context.Context, func(), bool) {
	if d.conn.IsClosed() {
		status.InternalServerError(c, nil)
		return nil, nil, func() {}, false
	}

	ctx := context.Background()

	tx, err := d.conn.Begin(ctx)
	if err != nil {
		status.InternalServerError(c, nil)
		return nil, nil, func() {}, false
	}

	return d.qry.WithTx(tx), &ctx, func() {
		commitTransactionIfSuccess(c, tx, &ctx)
	}, true
}

func commitTransactionIfSuccess(c *fiber.Ctx, tx pgx.Tx, ctx *context.Context) {
	if c.Response().StatusCode()/100 != 2 {
		tx.Rollback(*ctx)
		return
	}

	tx.Commit(*ctx)
}

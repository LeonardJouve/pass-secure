package database

import (
	"context"
	"embed"
	"errors"
	"path"

	"github.com/LeonardJouve/pass-secure/database/queries"
	"github.com/LeonardJouve/pass-secure/status"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Database struct {
	conn *pgx.Conn
	qry  *queries.Queries
	ctx  context.Context
}

const MIGRATIONS_FOLDER = "migrations"

//go:embed migrations/*.sql
var migrations embed.FS

var db *Database

func New(connectionURL string) (*Database, error) {
	oldDb, err := GetInstance()
	if err == nil {
		oldDb.Close()
	}

	ctx := context.Background()

	conn, err := pgx.Connect(ctx, connectionURL)
	if err != nil {
		return &Database{}, err
	}

	db = &Database{
		ctx:  ctx,
		conn: conn,
		qry:  queries.New(conn),
	}

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

func (d *Database) Exec(sql string, arguments ...any) (pgconn.CommandTag, error) {
	return d.conn.Exec(d.ctx, sql, arguments...)
}

func (d *Database) WaitForNotification() (*pgconn.Notification, error) {
	return d.conn.WaitForNotification(d.ctx)
}

func (d *Database) Migrate() error {
	if d.conn.IsClosed() {
		return errors.New("database connection closed")
	}

	migrationsPath := MIGRATIONS_FOLDER

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

		_, err = d.conn.Exec(d.ctx, string(content))
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

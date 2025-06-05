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
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	pool *pgxpool.Pool
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

	pool, err := pgxpool.New(ctx, connectionURL)
	if err != nil {
		return &Database{}, err
	}

	db = &Database{
		ctx:  ctx,
		pool: pool,
		qry:  queries.New(pool),
	}

	return db, nil
}

func GetInstance() (*Database, error) {
	if db == nil || db.pool.Ping(db.ctx) != nil {
		return nil, errors.New("no database connection")
	}

	return db, nil
}

func Acquire() (*pgxpool.Conn, func(), context.Context, error) {
	db, err := GetInstance()
	if err != nil {
		return nil, nil, nil, err
	}

	conn, err := db.pool.Acquire(db.ctx)
	if err != nil {
		return nil, nil, nil, err
	}

	return conn, func() {
		conn.Release()
	}, db.ctx, nil
}

func BeginTransaction(c *fiber.Ctx) (*queries.Queries, context.Context, func(), bool) {
	db, err := GetInstance()
	if err != nil {
		status.InternalServerError(c, nil)
		return nil, nil, nil, false
	}

	conn, err := db.pool.Acquire(db.ctx)
	if err != nil {
		status.InternalServerError(c, nil)
		return nil, nil, nil, false
	}

	ctx := context.Background()

	tx, err := conn.Begin(ctx)
	if err != nil {
		status.InternalServerError(c, nil)
		return nil, nil, nil, false
	}

	return db.qry.WithTx(tx), ctx, func() {
		commitTransactionIfSuccess(c, tx, ctx)
		conn.Release()
	}, true
}

func (d *Database) Close() {
	d.pool.Close()
}

func (d *Database) Migrate() error {
	conn, err := d.pool.Acquire(d.ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

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

		_, err = conn.Exec(d.ctx, string(content))
		if err != nil {
			return err
		}
	}

	return nil
}

func commitTransactionIfSuccess(c *fiber.Ctx, tx pgx.Tx, ctx context.Context) {
	if c.Response().StatusCode()/100 != 2 {
		tx.Rollback(ctx)
		return
	}

	tx.Commit(ctx)
}

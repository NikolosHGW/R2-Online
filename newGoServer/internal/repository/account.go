// Package repository provides database access via pgx/v5.
// All functions take a context and a pgx connection/pool.
// No ORM — raw SQL for predictable behaviour and easy debugging.
package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB is the shared database connection pool. Pass it around, don't use globals.
type DB = pgxpool.Pool

// Account mirrors the accounts table.
type Account struct {
	ID       int32
	Login    string
	Password string // bcrypt / argon2id hash
}

// AccountRepository provides data access for the accounts table.
type AccountRepository struct {
	db *DB
}

func NewAccountRepository(db *DB) *AccountRepository {
	return &AccountRepository{db: db}
}

// GetByLogin fetches an account by login name.
// Returns (nil, nil) when not found.
func (r *AccountRepository) GetByLogin(ctx context.Context, login string) (*Account, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, login, password FROM accounts WHERE login = $1 AND banned_at IS NULL`,
		login,
	)
	var a Account
	if err := row.Scan(&a.ID, &a.Login, &a.Password); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

// GetByID fetches an account by primary key.
func (r *AccountRepository) GetByID(ctx context.Context, id int32) (*Account, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, login, password FROM accounts WHERE id = $1`,
		id,
	)
	var a Account
	if err := row.Scan(&a.ID, &a.Login, &a.Password); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

// Create inserts a new account and returns its id.
func (r *AccountRepository) Create(ctx context.Context, login, hashedPassword string) (int32, error) {
	var id int32
	err := r.db.QueryRow(ctx,
		`INSERT INTO accounts (login, password) VALUES ($1, $2) RETURNING id`,
		login, hashedPassword,
	).Scan(&id)
	return id, err
}

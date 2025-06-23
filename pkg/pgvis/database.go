package pgvis

import (
	"database/sql"
)

// DB contains all database tables
type DB struct {
	Users *DBUsers

	db *sql.DB
}

func NewDB(db *sql.DB) *DB {
	return &DB{
		Users: NewDBUsers(db),
		db:    db,
	}
}

type DBUsers struct {
	db *sql.DB
}

func NewDBUsers(db *sql.DB) *DBUsers {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			"id" INTEGER NOT NULL,
			"telegram_id" INTEGER NOT NULL,
			"user_name" TEXT NOT NULL,
			"api_key" TEXT NOT NULL,
			PRIMARY KEY("id" AUTOINCREMENT, "telegram_id")
		);
	`
	if _, err := db.Exec(query); err != nil {
		panic(err)
	}

	return &DBUsers{
		db: db,
	}
}

func (db *DBUsers) List() ([]*User, error) {
	users := NewUsers()

	query := `SELECT * FROM users`
	if r, err := db.db.Query(query); err != nil {
		return users, err
	} else {
		defer r.Close()

		user := &User{}
		for r.Next() {
			err := r.Scan(
				&user.ID,
				&user.TelegramID, &user.UserName,
				&user.ApiKey,
			)
			if err != nil {
				return users, err
			}

			users = append(users, user)
		}
	}

	return users, nil
}

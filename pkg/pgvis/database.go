package pgvis

import (
	"database/sql"
	"errors"
	"fmt"
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
			"telegram_id" INTEGER NOT NULL,
			"user_name" TEXT NOT NULL,
			"api_key" TEXT NOT NULL,
			PRIMARY KEY("telegram_id")
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

	query := `SELECT * FROM users ORDER BY telegram_id ASC`
	r, err := db.db.Query(query)
	if err != nil {
		return users, err
	}

	defer r.Close()

	for r.Next() {
		user := &User{}

		err = r.Scan(&user.TelegramID, &user.UserName, &user.ApiKey)
		if err != nil {
			return users, err
		}

		users = append(users, user)
	}

	return users, nil
}

func (db *DBUsers) Get(telegramID int64) (*User, error) {
	query := fmt.Sprintf(`SELECT * FROM users WHERE telegram_id=%d`, telegramID)
	r, err := db.db.Query(query)
	if err != nil {
		return nil, err
	}

	defer r.Close()

	user := &User{}

	if !r.Next() {
		return nil, ErrNotFound
	}

	err = r.Scan(&user.TelegramID, &user.UserName, &user.ApiKey)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (db *DBUsers) Add(user *User) error {
	if user.TelegramID == 0 {
		return errors.New("Telegram ID cannot be 0")
	}

	query := fmt.Sprintf(
		`SELECT * FROM users WHERE telegram_id = %d OR user_name = "%s"`,
		user.TelegramID, user.UserName,
	)

	r, err := db.db.Query(query)
	if err != nil {
		return err
	}
	if r.Next() {
		return ErrAlreadyExists
	}
	r.Close()

	query = `INSERT INTO users (telegram_id, user_name, api_key) VALUES (?, "?", "?")`
	_, err = db.db.Exec(query, user.TelegramID, user.UserName, user.ApiKey)
	return err
}

func (db *DBUsers) Remove(telegramID int64) {
	// TODO: Remove table entries with `telegramID`
}

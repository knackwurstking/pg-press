package pgvis

import (
	"database/sql"
	"errors"
	"fmt"
)

// DB contains all database tables
type DB struct {
	Users   *DBUsers
	Cookies *DBCookies

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

func (db *DBUsers) GetUserFromApiKey(apiKey string) (*User, error) {
	query := fmt.Sprintf(`SELECT * FROM users WHERE api_key="%s"`, apiKey)
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

	ok := r.Next()
	r.Close()
	if ok {
		return ErrAlreadyExists
	}

	query = fmt.Sprintf(
		`INSERT INTO users (telegram_id, user_name, api_key) VALUES (%d, "%s", "%s")`,
		user.TelegramID, user.UserName, user.ApiKey,
	)

	_, err = db.db.Exec(query)
	return err
}

func (db *DBUsers) Remove(telegramID int64) {
	query := fmt.Sprintf(
		`DELETE FROM users WHERE telegram_id = "%d"`,
		telegramID,
	)

	_, _ = db.db.Exec(query)
}

func (db *DBUsers) Update(telegramID int64, user *User) error {
	query := fmt.Sprintf(`SELECT * FROM users WHERE telegram_id = %d`, telegramID)

	r, err := db.db.Query(query)
	if err != nil {
		return err
	}

	ok := r.Next()
	r.Close()
	if !ok {
		return ErrNotFound
	}

	query = fmt.Sprintf(
		`UPDATE users SET user_name = "%s", api_key = "%s" WHERE telegram_id = %d`,
		user.UserName, user.ApiKey, telegramID,
	)

	_, err = db.db.Exec(query)
	return err
}

type DBCookies struct {
	db *sql.DB
}

func NewDBCookies(db *sql.DB) *DBCookies {
	query := `
		CREATE TABLE IF NOT EXISTS cookies (
			user_agent TEXT NOT NULL,
            value TEXT NOT NULL,
			api_key TEXT NOT NULL,
			PRIMARY KEY("value")
		);
	`
	if _, err := db.Exec(query); err != nil {
		panic(err)
	}

	return &DBCookies{
		db: db,
	}
}

func (db *DBCookies) Get(value string) (*Cookie, error) {
	query := fmt.Sprintf(`SELECT * FROM cookies WHERE value="%s"`, value)
	r, err := db.db.Query(query)
	if err != nil {
		return nil, err
	}

	defer r.Close()

	cookie := &Cookie{}

	if !r.Next() {
		return nil, ErrNotFound
	}

	err = r.Scan(&cookie.UserAgent, &cookie.Value, &cookie.ApiKey)
	if err != nil {
		return nil, err
	}

	return cookie, nil
}

func (db *DBCookies) Add(cookie *Cookie) error {
	if cookie.Value == "" {
		return errors.New("cookie \"Value\" cannot be empty")
	}

	query := fmt.Sprintf(
		`SELECT * FROM cookies WHERE value = "%s"`,
		cookie.Value,
	)

	r, err := db.db.Query(query)
	if err != nil {
		return err
	}

	ok := r.Next()
	r.Close()
	if ok {
		return ErrAlreadyExists
	}

	query = fmt.Sprintf(
		`INSERT INTO cookies (user_agent, value, api_key) VALUES ("%s", "%s", "%s")`,
		cookie.UserAgent, cookie.Value, cookie.ApiKey,
	)

	_, err = db.db.Exec(query)
	return err
}

func (db *DBCookies) Remove(value string) {
	query := fmt.Sprintf(
		`DELETE FROM cookies WHERE value = "%s"`,
		value,
	)

	_, _ = db.db.Exec(query)
}

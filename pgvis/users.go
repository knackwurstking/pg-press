package pgvis

import (
	"database/sql"
	"errors"
	"fmt"
)

type Users struct {
	db    *sql.DB
	feeds *Feeds
}

func NewUsers(db *sql.DB, feeds *Feeds) *Users {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			"telegram_id" INTEGER NOT NULL,
			"user_name" TEXT NOT NULL,
			"api_key" TEXT NOT NULL UNIQUE,
			PRIMARY KEY("telegram_id")
		);
	`
	if _, err := db.Exec(query); err != nil {
		panic(err)
	}

	return &Users{
		db:    db,
		feeds: feeds,
	}
}

func (db *Users) List() ([]*User, error) {
	users := []*User{}

	query := `SELECT * FROM users`
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

func (db *Users) Get(telegramID int64) (*User, error) {
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

func (db *Users) GetUserFromApiKey(apiKey string) (*User, error) {
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

func (db *Users) Add(user *User) error {
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

	// Feed...
	if err == nil {
		feed := NewUserAddFeed(user.UserName)
		if err := db.feeds.Add(feed); err != nil {
			return fmt.Errorf("add new feed: %s", err.Error())
		}
	}

	return err
}

func (db *Users) Remove(telegramID int64) error {
	user, _ := db.Get(telegramID)

	query := fmt.Sprintf(
		`DELETE FROM users WHERE telegram_id = "%d"`,
		telegramID,
	)

	_, err := db.db.Exec(query)

	// Feed...
	if err == nil && user != nil {
		feed := NewUserRemoveFeed(user.UserName)
		if err := db.feeds.Add(feed); err != nil {
			return fmt.Errorf("add new feed: %s", err.Error())
		}
	}

	return err
}

func (db *Users) Update(telegramID int64, user *User) error {
	prevUser, _ := db.Get(telegramID)

	query := fmt.Sprintf(`SELECT * FROM users WHERE telegram_id=%d`, telegramID)

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
		`UPDATE users SET user_name="%s", api_key="%s" WHERE telegram_id=%d`,
		user.UserName, user.ApiKey, telegramID,
	)

	_, err = db.db.Exec(query)

	// Feed...
	if err == nil && prevUser != nil {
		if prevUser.UserName != user.UserName {
			feed := NewUserNameChangeFeed(prevUser.UserName, user.UserName)
			if err = db.feeds.Add(feed); err != nil {
				return fmt.Errorf("add new feed: %s", err.Error())
			}
		}
	}

	return err
}

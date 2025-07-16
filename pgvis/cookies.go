package pgvis

import (
	"database/sql"
	"errors"
	"fmt"
	"slices"
)

type Cookies struct {
	db *sql.DB
}

func NewCookies(db *sql.DB) *Cookies {
	query := `
		CREATE TABLE IF NOT EXISTS cookies (
			user_agent TEXT NOT NULL,
			value TEXT NOT NULL,
			api_key TEXT NOT NULL,
			last_login INTEGER NOT NULL,
			PRIMARY KEY("value")
		);
	`

	if _, err := db.Exec(query); err != nil {
		panic(err)
	}

	return &Cookies{
		db: db,
	}
}

func (db *Cookies) List() ([]*Cookie, error) {
	cookies := []*Cookie{}

	query := `SELECT * FROM cookies ORDER BY last_login ASC`
	r, err := db.db.Query(query)
	if err != nil {
		return cookies, err
	}

	defer r.Close()

	for r.Next() {
		cookie := &Cookie{}

		err = r.Scan(&cookie.UserAgent, &cookie.Value, &cookie.ApiKey, &cookie.LastLogin)
		if err != nil {
			return cookies, err
		}

		cookies = append(cookies, cookie)
	}

	return cookies, nil
}

func (db *Cookies) ListApiKey(apiKey string) ([]*Cookie, error) {
	cookies := []*Cookie{}

	query := fmt.Sprintf(`SELECT * FROM cookies WHERE api_key="%s"`, apiKey)
	r, err := db.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	for r.Next() {
		cookie := Cookie{}

		err = r.Scan(&cookie.UserAgent, &cookie.Value, &cookie.ApiKey, &cookie.LastLogin)
		if err != nil {
			return nil, err
		}

		cookies = append(cookies, &cookie)
	}

	return cookies, nil
}

func (db *Cookies) Get(value string) (*Cookie, error) {
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

	err = r.Scan(&cookie.UserAgent, &cookie.Value, &cookie.ApiKey, &cookie.LastLogin)
	if err != nil {
		return nil, err
	}

	return cookie, nil
}

func (db *Cookies) Add(cookie *Cookie) error {
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
		`INSERT INTO cookies (user_agent, value, api_key, last_login) VALUES ("%s", "%s", "%s", %d)`,
		cookie.UserAgent, cookie.Value, cookie.ApiKey, cookie.LastLogin,
	)

	_, err = db.db.Exec(query)
	return err
}

func (db *Cookies) Update(value string, cookie *Cookie) error {
	if cookie.Value == "" {
		return errors.New("cookie \"Value\" cannot be empty")
	}

	query := fmt.Sprintf(
		`UPDATE cookies SET user_agent="%s", value="%s", api_key="%s", last_login=%d WHERE value="%s"`,
		cookie.UserAgent, cookie.Value, cookie.ApiKey, cookie.LastLogin, value,
	)

	_, err := db.db.Exec(query)
	return err
}

func (db *Cookies) Remove(value string) error {
	query := fmt.Sprintf(
		`DELETE FROM cookies WHERE value = "%s"`,
		value,
	)

	_, err := db.db.Exec(query)
	return err
}

func (db *Cookies) RemoveApiKey(apiKey string) error {
	query := fmt.Sprintf(
		`DELETE FROM cookies WHERE api_key = "%s"`,
		apiKey,
	)

	_, err := db.db.Exec(query)
	return err
}

func SortCookies(cookies []*Cookie) []*Cookie {
	cookiesSorted := []*Cookie{}

outer:
	for _, c := range cookies {
		for i, sc := range cookiesSorted {
			if c.LastLogin > sc.LastLogin {
				cookiesSorted = slices.Insert(cookiesSorted, i, c)
				continue outer
			}
		}

		cookiesSorted = append(cookiesSorted, c)
	}

	return cookiesSorted
}

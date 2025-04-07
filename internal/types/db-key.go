package types

import "fmt"

type DBKey struct {
	Time   int    `json:"time"`
	User   string `json:"user"`
	UserID int    `json:"user_id"`
}

func (dbkey *DBKey) Key() string {
	return fmt.Sprintf("%d:%s:%d", dbkey.Time, dbkey.User, dbkey.UserID)
}

func (dbkey *DBKey) Is(k *DBKey) bool {
	return k.Key() == dbkey.Key()
}

func (dbkey *DBKey) IsKey(k string) bool {
	return k == dbkey.Key()
}

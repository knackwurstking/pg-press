package helper

import (
	"fmt"
	"net/http"

	"github.com/knackwurstking/pg-press/internal/common"
	"github.com/knackwurstking/pg-press/internal/errors"
	"github.com/knackwurstking/pg-press/internal/shared"
)

func GetUserForApiKey(db *common.DB, apiKey string) (user *shared.User, merr *errors.MasterError) {
	users, merr := db.User.User.List()
	if merr != nil {
		return user, merr
	}
	for _, u := range users {
		if u.ApiKey == apiKey {
			user = u
			break
		}
	}
	if user == nil {
		return user, errors.NewMasterError(
			fmt.Errorf("no user found for api key %s", shared.MaskString(apiKey)),
			http.StatusNotFound,
		)
	}
	return user, merr
}

func ListCookiesForApiKey(db *common.DB, apiKey string) (cookies []*shared.Cookie, merr *errors.MasterError) {
	user, merr := GetUserForApiKey(db, apiKey)
	if merr != nil {
		return cookies, merr
	}
	if user.ApiKey != apiKey {
		return cookies, errors.NewMasterError(
			fmt.Errorf("api key mismatch for user id %d", user.ID),
			http.StatusUnauthorized,
		)
	}

	cookies, merr = db.User.Cookie.List()
	if merr != nil {
		return cookies, merr
	}
	// in-place filtering
	i := 0
	for _, cookie := range cookies {
		if cookie.UserID == user.ID {
			cookies[i] = cookie
			i++
		}
	}
	return cookies[:i], nil
}

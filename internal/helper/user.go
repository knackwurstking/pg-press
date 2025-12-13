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

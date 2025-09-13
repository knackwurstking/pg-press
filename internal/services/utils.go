package services

import (
	"fmt"

	"github.com/knackwurstking/pgpress/pkg/models"
)

func createUserInfo(user *models.User) string {
	return fmt.Sprintf("%s (ID: %d)", user.Name, user.TelegramID)
}

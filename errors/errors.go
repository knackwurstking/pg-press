package errors

import "net/http"

func GetHTTPStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}

	if IsNotFoundError(err) {
		return http.StatusNotFound
	}

	if IsAlreadyExistsError(err) {
		return http.StatusConflict
	}

	if IsInvalidCredentialsError(err) {
		return http.StatusUnauthorized
	}

	return http.StatusInternalServerError
}

package controller

import (
	"errors"
	"net/http"
	"regexp"
	"strings"

	"app/constants"
	"app/webpojo"
)

// validateEmail using regexp for check is email valid. Bad practice always trust to front-end info
func validateEmail(email string) bool {
	Re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return Re.MatchString(email)
}

// validateRegisterInfo returns true if all the required form values are passed. Return validation error text.
func validateRegisterInfo(req *http.Request, userCreateRequest *webpojo.UserCreateReq, expectedRole int) (bool, string) {
	if userCreateRequest.FirstName == "" {
		return false, "first_name is empty"
	}

	if userCreateRequest.LastName == "" {
		return false, "last_name is empty"
	}

	if userCreateRequest.Email == "" {
		return false, "email is empty"
	}

	if !validateEmail(userCreateRequest.Email) {
		return false, "bad email format"
	}

	valid, err := validatePassword(userCreateRequest.Password)
	if !valid {
		return false, err
	}

	if userCreateRequest.UserRole < 1 || userCreateRequest.UserRole > 4 {
		return false, "user role is wrong"
	}

	if expectedRole != constants.DefaultRole && (int(userCreateRequest.UserRole) != expectedRole) {
		return false, "user role unexpected"
	}

	return true, ""
}

func validatePassword(password string) (bool, string) {
	if password == "" {
		return false, "password is empty"
	}

	if len(password) < 5 {
		return false, "password too short - must have at least 5 sumbols"
	}

	return true, ""
}

// trimUserRegisterInfo trim spaces in users input info
func trimUserRegisterInfo(userCreateRequest *webpojo.UserCreateReq) error {
	if userCreateRequest == nil {
		return errors.New("error while trim user register request: user info is nil")
	}

	userCreateRequest.Email = strings.TrimSpace(userCreateRequest.Email)
	userCreateRequest.FirstName = strings.TrimSpace(userCreateRequest.FirstName)
	userCreateRequest.LastName = strings.TrimSpace(userCreateRequest.LastName)
	userCreateRequest.Password = strings.TrimSpace(userCreateRequest.Password)

	return nil
}

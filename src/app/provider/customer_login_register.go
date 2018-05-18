package provider

import (
	"app/webpojo"
	"errors"
	"strconv"

	"app/constants"
	"app/model"
)

// GetUserByEmail return user model if exist, else error
func GetUserByEmail(email string) (*model.User, error) {
	if email == "" {
		return nil, errors.New("error while get user by email: email is empty")
	}

	user, err := model.UserByEmail(email)
	if err == model.ErrNoResult {
		return nil, model.ErrNoResult
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByID return user model if exist, else error
func GetUserProfileByID(userID string) (*model.User, error) {
	if userID == "" {
		return nil, errors.New("error while get user's info by id: id is empty")
	}

	intUserID, err := strconv.Atoi(userID)
	if err != nil {
		return nil, errors.New("error while get user's info by id: can't parse id")
	}

	if intUserID == 0 {
		return nil, errors.New("get user's info by id: id is zero")
	}

	user, err := model.UserByID(userID)
	if err == model.ErrNoResult {
		return nil, model.ErrNoResult
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// IsUserExist is user's email already register in DB
func IsUserExist(email string) (bool, error) {
	if email == "" {
		return false, errors.New("error while check is user exist: email is empty")
	}

	_, err := model.UserByEmail(email)
	if err == model.ErrNoResult {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

// RegisterNewCustomer add new customer to DB
func RegisterNewCustomer(regReq *webpojo.UserCreateReq, password string) error {
	if regReq == nil {
		return errors.New("error while register new customer: request is nil")
	}

	if password == "" {
		return errors.New("error while register new customer: password is empty")
	}

	err := model.UserCreateWithRole(regReq.FirstName, regReq.LastName, regReq.Email, password, constants.CustomerRole)
	if err != nil {
		return err
	}

	return nil
}

// RemoveCustomerByEmail remove user info from DB. Need for testing and utils
func RemoveCustomerByEmail(email string) error {
	if email == "" {
		return errors.New("error while remove customer by email: email is empty")
	}

	err := model.UserRemoveByEmail(email)
	if err != nil {
		return err
	}

	return nil
}

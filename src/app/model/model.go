package model

import (
	"database/sql"
	"errors"
	"strings"
)

var (
	// ErrCode is a config or an internal error
	ErrCode = errors.New("DB configuration is incorrect")
	// ErrNoResult is a not results error
	ErrNoResult = errors.New("Result not found")
	// ErrUnavailable is a database not available error
	ErrUnavailable = errors.New("Database is unavailable")
	// ErrUnauthorized is a permissions violation
	ErrUnauthorized = errors.New("User does not have permission to perform this operation")
	// ErrConstraintFails return is app fails DB constraint
	ErrConstraintFails = errors.New("Constraint Fails")
	// ErrUserNotExist return if user not exist
	ErrUserNotExist = errors.New("User not exist")
	// ErrThreadNotExist return if thread not exist
	ErrThreadNotExist = errors.New("Thread not exist")
	// ErrMessageNotExist return if message not exist
	ErrMessageNotExist = errors.New("Message not exist")
)

// standardizeErrors returns the same error regardless of the database used
func standardizeError(err error) error {
	if err == nil {
		return nil
	}

	if err == sql.ErrNoRows {
		return ErrNoResult
	}

	if strings.HasPrefix(err.Error(), "Error 1452:") {
		return ErrConstraintFails
	}

	return err
}

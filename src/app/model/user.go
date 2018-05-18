package model

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"app/shared/database"
)

// *****************************************************************************
// User
// *****************************************************************************

// User table contains the information for each user
type User struct {
	ID        uint32    `db:"id"`
	FirstName string    `db:"first_name"`
	LastName  string    `db:"last_name"`
	Email     string    `db:"email"`
	Password  string    `db:"password"`
	StatusID  uint8     `db:"status_id"`
	UserRole  uint8     `db:"user_role"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	Deleted   uint8     `db:"deleted"`
}

// UserStatus table contains every possible user status (active/inactive)
type UserStatus struct {
	ID        uint8     `db:"id"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	Deleted   uint8     `db:"deleted"`
}

// UserID returns the user id
func (u *User) UserID() string {
	r := ""

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		r = fmt.Sprintf("%v", u.ID)
	}

	return r
}

// UserByID gets user information by ID
func UserByID(userID string) (User, error) {
	var err error

	result := User{}

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		err = database.SQL.Get(&result, "SELECT id, email, password, status_id, first_name, last_name, user_role FROM user WHERE id = ? LIMIT 1", userID)
	default:
		err = ErrCode
	}

	return result, standardizeError(err)
}

// UserByEmail gets user information from email
func UserByEmail(email string) (User, error) {
	var err error

	result := User{}

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		err = database.SQL.Get(&result, "SELECT id, email, password, status_id, first_name, last_name, user_role FROM user WHERE email = ? LIMIT 1", email)
	default:
		err = ErrCode
	}
	//fmt.Printf("%s", result)

	return result, standardizeError(err)
}

// UserCreate creates user
func UserCreate(firstName, lastName, email, password string) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		_, err = database.SQL.Exec("INSERT INTO user (first_name, last_name, email, password) VALUES (?,?,?,?)", firstName,
			lastName, email, password)

	default:
		err = ErrCode
	}

	return standardizeError(err)
}

// UserCreateWithRole creates user
func UserCreateWithRole(firstName, lastName, email, password string, userRole int) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		_, err = database.SQL.Exec("INSERT INTO user (first_name, last_name, email, password, user_role) VALUES (?,?,?,?,?)", firstName,
			lastName, email, password, userRole)

	default:
		err = ErrCode
	}

	return standardizeError(err)
}

//UserUpdate updates an user
func UserUpdate(user User) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		_, err = database.SQL.Exec("UPDATE user SET email=?, password=?, first_name=?, last_name=?, user_role=? "+
			"WHERE id=?  LIMIT 1", user.Email, user.Password, user.FirstName, user.LastName, user.UserRole, user.ID)

	default:
		err = ErrCode
	}

	return standardizeError(err)
}

// UserRemoveByEmail removing user from DB
func UserRemoveByEmail(email string) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		_, err = database.SQL.Exec("DELETE FROM user where email = ?", email)

	default:
		err = ErrCode
	}

	return standardizeError(err)
}

// CheckIsUserExist check is user exist or return error
func CheckIsUserExist(userID uint32) (bool, error) {
	var exists bool
	err := database.SQL.QueryRow("SELECT exists (SELECT id FROM user WHERE id=?)", userID).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return false, errors.New("error checking if row exists: " + err.Error())
	}
	return exists, nil
}

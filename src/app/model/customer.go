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

// Customer model contains all customer info
type Customer struct {
	ID             sql.NullInt64  `db:"id"`
	UserID         sql.NullInt64  `db:"user_id"`
	FirstName      string         `db:"first_name"`
	LastName       string         `db:"last_name"`
	Email          string         `db:"email"`
	UserName       sql.NullString `db:"user_name"`
	MailingAddress sql.NullString `db:"mailing_address"`
	Phone          sql.NullString `db:"phone"`
	Password       string         `db:"password"`
	StatusID       uint8          `db:"status_id"`
	UserRole       uint8          `db:"user_role"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
	Deleted        uint8          `db:"deleted"`
}

// RawCustomer model contains unique info for customer
type RawCustomer struct {
	ID             uint32 `db:"id"`
	UserID         uint32 `db:"user_id"`
	UserName       string `db:"user_name"`
	MailingAddress string `db:"mailing_address"`
	Phone          string `db:"phone"`
}

// CustomerID returns the customer id
func (c *Customer) CustomerID() string {
	r := ""

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		r = fmt.Sprintf("%v", c.ID)
	}

	return r
}

// RawCustomerByUserID gets raw customer info by user ID
func RawCustomerByUserID(userID string) (*RawCustomer, error) {
	var err error

	result := RawCustomer{}

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		err = database.SQL.Get(&result, `SELECT 
			id,
			user_id,
			user_name, 
			mailing_address, 
			phone 
			FROM customer_profile_info 
			WHERE user_id = ? LIMIT 1`, userID)
	default:
		err = ErrCode
	}

	return &result, standardizeError(err)
}

// CustomerByUserID gets user information by ID
func CustomerByUserID(userID string) (*Customer, error) {
	var err error

	result := Customer{}

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		err = database.SQL.Get(&result, `SELECT 
			cpi.id,
			cpi.user_id,
			u.password, 
			u.status_id, 
			u.first_name, 
			u.last_name, 
			u.user_role, 
			u.email,
			cpi.user_name, 
			cpi.mailing_address, 
			cpi.phone 
			FROM user u
			LEFT JOIN customer_profile_info cpi  ON cpi.user_id = u.id
			WHERE u.id = ? LIMIT 1`, userID)
	default:
		err = ErrCode
	}

	return &result, standardizeError(err)
}

//CustomerByEmail gets customer information from email
func CustomerByEmail(email string) (*Customer, error) {
	var err error

	result := Customer{}

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		err = database.SQL.Get(&result, `SELECT 
			cpi.id,
			cpi.user_id,
			u.password, 
			u.status_id, 
			u.first_name, 
			u.last_name, 
			u.user_role, 
			cpi.user_name, 
			cpi.mailing_address, 
			cpi.phone  
			FROM user u
			LEFT JOIN customer_profile_info cpi  ON cpi.user_id = u.id
			WHERE u.email = ? LIMIT 1`, email)
	default:
		err = ErrCode
	}

	return &result, standardizeError(err)
}

//CustomerUpdate updates customers info
func CustomerUpdate(customer *Customer) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		tx, err := database.SQL.Begin()
		if err != nil {
			return errors.New("error while start transaction: " + err.Error())
		}

		_, err = tx.Exec("UPDATE user SET email=?, password=?, first_name=?, last_name=? "+
			"WHERE id=?  LIMIT 1", customer.Email, customer.Password, customer.FirstName, customer.LastName, customer.UserID)
		if err != nil {
			tx.Rollback()
			return errors.New("error while transact (update) user's part of customer data: " + err.Error())
		}

		_, err = tx.Exec("UPDATE customer_profile_info SET user_name=?, mailing_address=?, phone=? "+
			"WHERE user_id=?  LIMIT 1", customer.UserName, customer.MailingAddress, customer.Phone, customer.UserID)
		if err != nil {
			tx.Rollback()
			return errors.New("error while transact customer's part of user data: " + err.Error())
		}

		err = tx.Commit()
		if err != nil {
			tx.Rollback()
			return errors.New("errro while commit transaction")
		}

	default:
		err = ErrCode
	}

	return standardizeError(err)
}

//CustomerInsert insert new customer profile info
func CustomerInsert(customer *Customer) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		tx, err := database.SQL.Begin()
		if err != nil {
			return errors.New("error while start transaction: " + err.Error())
		}
		_, err = tx.Exec("UPDATE user SET email=?, password=?, first_name=?, last_name=? "+
			"WHERE id=?  LIMIT 1", customer.Email, customer.Password, customer.FirstName, customer.LastName, customer.UserID)
		if err != nil {
			tx.Rollback()
			return errors.New("error while transact (insert) user's part of customer data: " + err.Error())
		}

		_, err = tx.Exec("INSERT INTO customer_profile_info (user_id, user_name, mailing_address, phone) VALUES(?,?,?,?)", customer.UserID, customer.UserName, customer.MailingAddress, customer.Phone)
		if err != nil {
			tx.Rollback()
			return errors.New("error while transact user's part of customer data: " + err.Error())
		}

		err = tx.Commit()
		if err != nil {
			tx.Rollback()
			return errors.New("errro while commit transaction")
		}

	default:
		err = ErrCode
	}

	return standardizeError(err)
}

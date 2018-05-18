package model

import (
	"app/webpojo"
	"errors"
	"fmt"
	"log"
	"time"

	"app/shared/database"
)

// *****************************************************************************
// Lender CRUD
// *****************************************************************************

// Lender table contains the information for each lead
type Lender struct {
	ID          uint32    `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	Address     string    `db:"address"`
	Email       string    `db:"email"`
	Phone       string    `db:"phone"`
	Contact     string    `db:"contact"`
	StatusID    int       `db:"status_id"`
	CreatedAt   time.Time `db:"created_at"`
}

// LenderCreate inserts a new lender
func LenderCreate(name string, description string, address string, email string,
	phone string, contact string, statusId int) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		result := false
		retries := 0
		for !result && retries < 5 {
			_, err = database.SQL.Exec("INSERT INTO lender (name, description, address, email, phone, contact, status_id)"+
				" VALUES (?,?,?,?,?,?,?)",
				name, description, address, email, phone, contact, statusId)

			if err == nil {
				result = true
				log.Println("Inserted lender " + name + " successfully")
			} else {
				log.Println(fmt.Sprintf("%v has error %v", name, err))
			}
			retries = retries + 1
		}
	default:
		err = ErrCode
	}

	return standardizeError(err)
}

// LenderUpdate updates an existing lender
func LenderUpdate(name string, description string, address string, email string,
	phone string, contact string, statusId int) error {
	var err error
	log.Println("Status: " + string(statusId))
	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		result := false
		retries := 0
		for !result && retries < 5 {
			_, err = database.SQL.Exec("UPDATE lender SET description = ?, address = ?, email = ?, phone = ?, contact = ?, "+
				"status_id = ? WHERE name = ?", description, address, email, phone, contact, statusId, name)

			if err == nil {
				result = true
				log.Println("Updated lender " + name + " successfully")
			} else {
				log.Println(fmt.Sprintf("%v has error %v", name, err))
			}
			retries = retries + 1
		}
	default:
		err = ErrCode
	}

	return standardizeError(err)
}

//LeaderByName gets Lender information from leader name
func LeaderByName(lenderName string) (Lender, error) {
	var err error
	result := Lender{}

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		err = database.SQL.Get(&result, "SELECT id, name, description, address, email, phone, "+
			"contact, status_id, created_at FROM lender WHERE name = ? LIMIT 1", lenderName)
	default:
		err = ErrCode
	}
	fmt.Printf("Email: %s\n", result.Email)

	if err != nil {
		fmt.Printf("%s", err)
	}

	return result, standardizeError(err)
}

// LendersListAll gets all lenders
func LendersListAll() ([]Lender, error) {
	var err error
	var lenders []Lender

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		err = database.SQL.Select(&lenders, "SELECT id, name, description, address, email, phone, contact, status_id, created_at from lender")
	default:
		err = ErrCode
	}

	return lenders, standardizeError(err)
}

// LenderRatesListBest gets best lender rates (min apr+interest from group of products)
func LenderRatesListBest() ([]LenderRate, error) {
	var err error
	var lenderRates []LenderRate

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		err = database.SQL.Select(&lenderRates,
			`SELECT id,
			lender_id,
			lender_name,
			product,
			interest,
			apr,
			begin_date,
			end_date,
			created_at 
				FROM lender_rate WHERE (interest, product) IN (
                SELECT
				max(r2.interest), 
                r2.product
					FROM lender_rate r2 
					GROUP BY r2.product)
                GROUP BY product`)
	default:
		err = ErrCode
	}

	return lenderRates, standardizeError(err)
}

// FastQuoteRate return info for fast quote
func FastQuoteRate(fastQuoteRequest *webpojo.FastQuoteReq) ([]*LenderRate, error) {
	var err error
	lenderRate := []*LenderRate{}

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		err = database.SQL.Select(&lenderRate,
			`SELECT id,
			lender_id,
			lender_name,
			product,
			interest,
			apr,
			begin_date,
			end_date,
			created_at 
				FROM lender_rate WHERE (interest, product) IN (
                SELECT
				max(r2.interest), 
                r2.product
					FROM lender_rate r2 
					GROUP BY r2.product)
				GROUP BY product;`)
	default:
		err = ErrCode
	}

	if len(lenderRate) == 0 {
		return nil, standardizeError(errors.New("lenders list is empty"))
	}

	return lenderRate, standardizeError(err)
}

//LenderDeleteByName deletes a lender by its name
func LenderDeleteByName(lenderName string) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		_, err = database.SQL.Exec("DELETE FROM lender WHERE name = ?", lenderName)
	default:
		err = ErrCode
	}

	if err != nil {
		fmt.Printf("%s", err)
	}

	return standardizeError(err)
}

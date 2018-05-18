package model

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"app/shared/database"
)

// *****************************************************************************
// Note
// *****************************************************************************

// UserFile table contains the information for each file
type UserFile struct {
	ID        uint32    `db:"id" bson:"id,omitempty"` // Don't use Id, use NoteID() instead for consistency with MongoDB
	FileName  string    `db:"file_name" bson:"file_name"`
	FileType  string    `db:"file_type" bson:"file_type"`
	UserID    uint32    `db:"user_id"`
	CreatedAt time.Time `db:"created_at" bson:"created_at"`
	UpdatedAt time.Time `db:"updated_at" bson:"updated_at"`
	Deleted   uint8     `db:"deleted" bson:"deleted"`
}

// FileID returns the file id
func (us *UserFile) FileID() string {
	r := ""

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		r = fmt.Sprintf("%v", us.ID)
	}

	return r
}

// FileByID gets note by ID
func FileByID(userID string, fileID string) (*UserFile, error) {
	var err error

	result := UserFile{}

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		err = database.SQL.Get(&result, "SELECT id, file_name, file_type, user_id, created_at, updated_at, deleted FROM uploaded_files WHERE id = ? AND user_id = ? LIMIT 1", fileID, userID)
	default:
		err = ErrCode
	}

	return &result, standardizeError(err)
}

// FilesByUserID gets all files for a user
func FilesByUserID(userID string) ([]*UserFile, error) {
	var err error

	var result []*UserFile

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		err = database.SQL.Select(&result, "SELECT id, file_name, file_type, user_id, created_at, updated_at, deleted FROM uploaded_files WHERE user_id = ?", userID)
	default:
		err = ErrCode
	}

	return result, standardizeError(err)
}

// FileCreate creates a new file in DB
func FileCreate(fileName, fileType, userID string) (int, error) {
	var err error
	var res sql.Result
	var lastID int64

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		res, err = database.SQL.Exec("INSERT INTO uploaded_files (file_name, file_type, user_id) VALUES (?,?,?)", fileName, fileType, userID)

	default:
		err = ErrCode
	}

	if err != nil {
		return 0, standardizeError(err)
	}

	lastID, err = res.LastInsertId()
	if err != nil {
		return 0, standardizeError(err)
	}

	return int(lastID), standardizeError(err)
}

// FileDelete deletes a file
func FileDelete(userID string, fileID string) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		res, err := database.SQL.Exec("DELETE FROM uploaded_files WHERE id = ? AND user_id = ?", fileID, userID)
		if err != nil {
			return standardizeError(err)
		}

		if res == nil {
			log.Println("file delete result is nil")
			return standardizeError(err)
		}

		rows, err := res.RowsAffected()
		if err != nil {
			log.Println("error while get rows affected: " + err.Error())
			return standardizeError(err)
		}

		if rows == 0 {
			return ErrNoResult
		}
	default:
		err = ErrCode
	}

	return standardizeError(err)
}

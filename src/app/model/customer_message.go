package model

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"app/shared/database"
)

// *****************************************************************************
// Customer message
// *****************************************************************************

// Message table contains the information for each user message
type Message struct {
	ID                uint32    `db:"id"`
	ThreadID          uint32    `db:"thread_id"`
	ToUserID          uint32    `db:"to_user_id"`
	ToUserFirstName   string    `db:"to_first_name"`
	ToUserLastName    string    `db:"to_last_name"`
	FromUserID        uint32    `db:"from_user_id"`
	FromUserFirstName string    `db:"from_first_name"`
	FromUserLastName  string    `db:"from_last_name"`
	Content           string    `db:"content"`
	Readed            bool      `db:"readed"`
	CreatedAt         time.Time `db:"created_at"`
	UpdatedAt         time.Time `db:"updated_at"`
	Deleted           uint8     `db:"deleted"`
}

// MessageID returns the message id
func (m *Message) MessageID() string {
	r := ""

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		r = fmt.Sprintf("%v", m.ID)
	}

	return r
}

// MessageByID gets message by ID
func MessageByID(toUserID string, messageID string) (*Message, error) {
	var err error

	result := &Message{}

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		err = database.SQL.Get(result, `
			SELECT customer_message.id, 
            customer_message.thread_id,
			to_user_id,
 			to_user_info.first_name as to_first_name, 
 			to_user_info.last_name as to_last_name, 
 		
			from_user_id, 
 			from_user_info.first_name as from_first_name, 
			from_user_info.last_name as from_last_name, 
			
            content,
            readed,
				
			customer_message.created_at, 
 			customer_message.updated_at, 
 			customer_message.deleted 
 
			FROM customer_message 

			LEFT JOIN user from_user_info ON from_user_info.id=from_user_id
			LEFT JOIN user to_user_info ON to_user_info.id=to_user_id 

			WHERE customer_message.id = ? AND to_user_id = ? LIMIT 1;
			`, messageID, toUserID)
	default:
		err = ErrCode
	}

	return result, standardizeError(err)
}

// MessagesByUserID gets messages for a user
func MessagesByUserID(threadID, userID uint32, count, offset int) ([]*Message, error) {
	log.Println(threadID, " ", userID, " ", count, " ", offset)
	var err error

	var result []*Message

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		err = database.SQL.Select(&result, `
			SELECT customer_message.id, 
			
            customer_message.thread_id,
			to_user_id,
 			to_user_info.first_name as to_first_name, 
 			to_user_info.last_name as to_last_name, 
 		
			from_user_id, 
 			from_user_info.first_name as from_first_name, 
			from_user_info.last_name as from_last_name, 
			
            content,
            readed,
				
			customer_message.created_at, 
 			customer_message.updated_at, 
 			customer_message.deleted 
 
			FROM customer_message 

			LEFT JOIN user from_user_info ON from_user_info.id=from_user_id
			LEFT JOIN user to_user_info ON to_user_info.id=to_user_id 

			WHERE thread_id = ? AND (to_user_id=? OR from_user_id=?) ORDER BY created_at DESC LIMIT ? OFFSET ?;`, threadID, userID, userID, count, offset)
	default:
		err = ErrCode
	}

	return result, standardizeError(err)
}

// MessageCreate creates a message
func MessageCreate(toUserID, fromUserID, threadID uint32, content string) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		_, err = database.SQL.Exec("INSERT INTO customer_message (to_user_id, from_user_id, content, thread_id) VALUES (?, ?, ?, ?);", toUserID, fromUserID, content, threadID)
		if err != nil {
			return standardizeError(err)
		}

	default:
		err = ErrCode
	}

	return standardizeError(err)
}

// MessageUpdate updates a message
func MessageUpdate(content string, fromUserID, messageID, threadID uint32) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		_, err := database.SQL.Exec("UPDATE customer_message SET content=?  WHERE id = ? AND thread_id = ? AND from_user_id = ? LIMIT 1", content, messageID, threadID, fromUserID)
		if err != nil {
			return standardizeError(err)
		}

	default:
		err = ErrCode
	}

	return standardizeError(err)
}

// MarkMessageReadedUpdate set messages readed mark
func MarkMessageReadedUpdate(readed bool, toUserID, messageID uint32, threadID uint32) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		res, err := database.SQL.Exec("UPDATE customer_message SET readed=?  WHERE id = ? AND thread_id = ? AND to_user_id = ? LIMIT 1", readed, messageID, threadID, toUserID)
		if err != nil {
			return standardizeError(err)
		}

		affected, err := res.RowsAffected()
		if err != nil {
			log.Println("error while mark message readed: ", err)
			return standardizeError(err)
		}

		if affected == 0 {
			log.Println("error while mark message readed: 0 rows affected")
			return standardizeError(err)
		}

	default:
		err = ErrCode
	}

	return standardizeError(err)
}

// MessageDelete deletes a message
func MessageDelete(fromUserID, messageID, threadID uint32) error {
	var err error

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		_, err = database.SQL.Exec("DELETE FROM customer_message WHERE id = ? AND thread_id = ? AND from_user_id = ?", messageID, threadID, fromUserID)

	default:
		err = ErrCode
	}

	return standardizeError(err)
}

// CheckIsMessageExist check is user exist or return error
func CheckIsMessageExist(messageID uint32) (bool, error) {
	var exists bool
	err := database.SQL.QueryRow("SELECT exists (SELECT id FROM customer_message WHERE id=?)", messageID).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return false, errors.New("error checking if row exists: " + err.Error())
	}
	return exists, nil
}

// CheckIsThreadExist check is messages thread exist
func CheckIsThreadExist(threadID uint32) (bool, error) {
	var exists bool
	err := database.SQL.QueryRow("SELECT exists (SELECT id FROM message_thread WHERE id=?)", threadID).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return false, errors.New("error checking if row exists: " + err.Error())
	}
	return exists, nil
}

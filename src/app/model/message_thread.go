package model

import (
	"app/shared/database"
	"log"
	"time"
)

// *****************************************************************************
// Customer message thread
// *****************************************************************************

// MessageThread represent users message series to each other
type MessageThread struct {
	ID         uint32    `db:"id"`
	ToUserID   uint32    `db:"to_user_id"`
	FromUserID uint32    `db:"from_user_id"`
	Title      string    `db:"title"`
	Content    string    `db:"content"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
	Deleted    uint8     `db:"deleted"`
}

// ThreadsByUserID return threads info for
func ThreadsByUserID(toUserID string, limit, offset int) ([]*MessageThread, error) {
	var err error

	var result []*MessageThread

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		err = database.SQL.Select(&result, `
	SELECT message_thread.id, 
	message_thread.to_user_id, 
	message_thread.from_user_id, 
	message_thread.title,
	last_message.content,
	message_thread.created_at, 
	message_thread.updated_at, 
	message_thread.deleted

	FROM message_thread
	JOIN (select * from customer_message where `+"`created_at`"+` in (
	select max(`+"`created_at`"+`) 
	from customer_message
	group by thread_id)) last_message ON last_message.thread_id=message_thread.id

	WHERE message_thread.to_user_id = ? OR message_thread.from_user_id = ?
	GROUP BY id
	ORDER BY created_at DESC LIMIT ? OFFSET ?`, toUserID, toUserID, limit, offset)
	default:
		err = ErrCode
	}

	return result, standardizeError(err)
}

// ThreadCreate create new messages thread
func ThreadCreate(title string, toUserID, fromUserID uint32) (uint32, error) {
	var err error
	var lastID int64

	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		result, err := database.SQL.Exec("INSERT INTO message_thread (to_user_id, from_user_id, title) VALUES (?,?,?)", toUserID, fromUserID, title)
		if err != nil {
			return 0, standardizeError(err)
		}

		lastID, err = result.LastInsertId()
		if err != nil {
			log.Println("error while get last inserted id: ", err)
			return 0, standardizeError(err)
		}

	default:
		err = ErrCode
	}

	return uint32(lastID), standardizeError(err)
}

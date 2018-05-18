package webpojo

// PostNewMessageResp return info about thread after message sending
type PostNewMessageResp struct {
	NewThreadCreated bool   `json:"new_thread_created"`
	ThreadID         uint32 `json:"thread_id"`
}

// MessagesThreadsPostReq contains messages threads request param
type MessagesThreadsPostReq struct {
	Count  int `json:"count"`
	Offset int `json:"offseet"`
}

// MessagesThreadsResp represents user's messages threads
type MessagesThreadsResp struct {
	ID         uint32 `json:"id"`
	ToUserID   uint32 `json:"to_user_id"`
	FromUserID uint32 `json:"from_user_id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	CreatedAt  string `json:"created_at"`
}

// MessagesListPostReq contains params for read messages
type MessagesListPostReq struct {
	ThreadID uint32 `json:"thread_id"`
	Count    int    `json:"count"`
	Offset   int    `json:"offseet"`
}

// MessagePostReq contain info for creating new message for user
type MessagePostReq struct {
	ToUserID   uint32 `json:"to_user_id"`
	FromUserID uint32 `json:"from_user_id"`
	Subject    string `json:"subject"`
	Content    string `json:"content"`
	ThreadID   uint32 `json:"thread_id"`
}

// MessagePatchReq patch user's message
type MessagePatchReq struct {
	ToUserID   uint32 `json:"to_user_id"`
	FromUserID uint32 `json:"from_user_id"`
	MessageID  uint32 `json:"message_id"`
	Content    string `json:"content"`
	ThreadID   uint32 `json:"thread_id"`
}

// UserMessagePatchReq using for mark message raded
type UserMessagePatchReq struct {
	MessageID uint32 `json:"message_id"`
	Readed    bool   `json:"readed"`
	ThreadID  uint32 `json:"thread_id"`
}

// MessageDeleteReq request model for deleting message
type MessageDeleteReq struct {
	ToUserID   uint32 `json:"to_user_id"`
	FromUserID uint32 `json:"from_user_id"`
	MessageID  uint32 `json:"message_id"`
	ThreadID   uint32 `json:"thread_id"`
}

// MessageListResp contain info for one user's message
type MessageListResp struct {
	ThreadID   uint32 `json:"thread_id"`
	MessageID  uint32 `json:"message_id"`
	FromUserID uint32 `json:"from_id"`
	FromFirst  string `json:"from_first"`
	FromLast   string `json:"from_last"`
	ToUserID   uint32 `json:"to_id"`
	ToFirst    string `json:"to_first"`
	ToLast     string `json:"to_last"`
	Subject    string `json:"subject,omitempty"`
	Content    string `json:"content"`
	Readed     bool   `json:"readed"`
	Created    string `json:"created"`
}

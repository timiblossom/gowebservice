package provider

import (
	"app/model"
	"app/webpojo"
	"errors"
	"log"
	"strconv"
)

// ThreadsByUserID return list of user's threads
func ThreadsByUserID(userID string, count, offset int) ([]*webpojo.MessagesThreadsResp, error) {
	if userID == "" {
		return nil, errors.New("error while get user's threads: user ID is empty")
	}

	if count <= 0 {
		return nil, errors.New("error while get user's threads: count is 0")
	}

	if offset < 0 {
		return nil, errors.New("error while get user's threads: offset less than 0")
	}

	threads, err := model.ThreadsByUserID(userID, count, offset)
	if err != nil {
		log.Println("error while get user's threads: ", err)
		return nil, err
	}

	res := []*webpojo.MessagesThreadsResp{}

	for _, v := range threads {
		res = append(res, &webpojo.MessagesThreadsResp{
			ID:         v.ID,
			ToUserID:   v.ToUserID,
			FromUserID: v.FromUserID,
			Title:      v.Title,
			Content:    v.Content,
			CreatedAt:  v.CreatedAt.String(),
		})
	}

	return res, nil
}

// PostNewMessage create new user's message in database
func PostNewMessage(userID string, messagePostReq *webpojo.MessagePostReq) (*webpojo.PostNewMessageResp, error) {
	var newThreadCreated bool

	err := CheckUserExist(messagePostReq.FromUserID)
	if err != nil {
		log.Println("error while check is user exist: ", err)
		return nil, err
	}

	err = CheckUserExist(messagePostReq.ToUserID)
	if err != nil {
		log.Println("error while check is user exist: ", err)
		return nil, err
	}

	if messagePostReq.ThreadID != 0 {
		err := CheckThreadExist(messagePostReq.ThreadID)
		if err != nil {
			return nil, err
		}
	} else {
		messagePostReq.ThreadID, err = CreateNewThread(messagePostReq.FromUserID, messagePostReq.ToUserID, messagePostReq.Subject)
		if err != nil {
			return nil, err
		}
		newThreadCreated = true
	}

	err = model.MessageCreate(messagePostReq.ToUserID, messagePostReq.FromUserID, messagePostReq.ThreadID, messagePostReq.Content)
	if err != nil {
		log.Println("error while create new user's message: " + err.Error())
		return nil, err
	}

	return &webpojo.PostNewMessageResp{
		NewThreadCreated: newThreadCreated,
		ThreadID:         messagePostReq.ThreadID,
	}, nil
}

// CheckThreadExist return error if thread not exist
func CheckThreadExist(threadID uint32) error {
	threadExist, err := model.CheckIsThreadExist(threadID)
	if err != nil {
		log.Println("error while check is thread exist: " + err.Error())
		return err
	}

	if !threadExist {
		log.Println("thread not exist")
		return model.ErrThreadNotExist
	}

	return nil
}

// CheckMessageExist return error if thread not exist
func CheckMessageExist(messageID uint32) error {
	threadExist, err := model.CheckIsMessageExist(messageID)
	if err != nil {
		log.Println("error while check is message exist: " + err.Error())
		return err
	}

	if !threadExist {
		log.Println("message not exist")
		return model.ErrMessageNotExist
	}

	return nil
}

// CheckUserExist return error if user not exist
func CheckUserExist(userID uint32) error {
	userExist, err := model.CheckIsUserExist(userID)
	if err != nil {
		log.Println("error while check is user exist: " + err.Error())
		return err
	}

	if !userExist {
		log.Println("user not exist")
		return model.ErrUserNotExist
	}

	return nil
}

// CreateNewThread chack param and create new thread witj messages
func CreateNewThread(fromUserID, toUserID uint32, title string) (uint32, error) {
	log.Println(fromUserID, " - ", toUserID, " - ", title)
	if fromUserID == 0 {
		log.Println("error while create new thread: from user ID is 0")
		return 0, errors.New("from user ID is 0")
	}

	if toUserID == 0 {
		log.Println("error while create new thread: to user ID is 0")
		return 0, errors.New("to user ID is 0")
	}

	threadID, err := model.ThreadCreate(title, toUserID, fromUserID)
	if err != nil {
		log.Println("error while create new thread: ", err)
		return 0, err
	}

	return threadID, nil
}

// PatchMessage path user's message in database
func PatchMessage(messagePatchReq *webpojo.MessagePatchReq) error {
	err := CheckUserExist(messagePatchReq.ToUserID)
	if err != nil {
		log.Println("error while check is user exist: ", err)
		return err
	}

	err = CheckUserExist(messagePatchReq.FromUserID)
	if err != nil {
		log.Println("error while check is user exist: ", err)
		return err
	}

	err = CheckMessageExist(messagePatchReq.MessageID)
	if err != nil {
		log.Println("error while check is message exist: ", err)
		return err
	}

	err = CheckThreadExist(messagePatchReq.ThreadID)
	if err != nil {
		log.Println("error while check is thread exist: ", err)
		return err
	}

	err = model.MessageUpdate(messagePatchReq.Content, messagePatchReq.FromUserID, messagePatchReq.MessageID, messagePatchReq.ThreadID)
	if err != nil {
		log.Println("error while patch user's message: " + err.Error())
		return err
	}

	return nil
}

// DeleteMessage delete users message that already send
func DeleteMessage(fromUserID, messageID, threadID uint32) error {
	err := CheckThreadExist(threadID)
	if err != nil {
		log.Println("error while delete message: error while check user exist: ", err)
		return model.ErrThreadNotExist
	}

	err = CheckUserExist(fromUserID)
	if err != nil {
		log.Println("error while delete message: error while check is user exist: " + err.Error())
		return model.ErrUserNotExist
	}

	err = CheckMessageExist(messageID)
	if err != nil {
		log.Println("error while delete message: error while check is message exist: " + err.Error())
		return model.ErrMessageNotExist
	}

	err = model.MessageDelete(fromUserID, messageID, threadID)
	if err != nil {
		log.Println("error while patch user's message: " + err.Error())
		return err
	}

	return nil
}

// MarkMessageReaded path user's message in database
func MarkMessageReaded(userID string, messagePatchReq *webpojo.UserMessagePatchReq) error {
	err := CheckThreadExist(messagePatchReq.ThreadID)
	if err != nil {
		log.Println("error while mark message readed: " + err.Error())
		return model.ErrThreadNotExist
	}

	err = CheckMessageExist(messagePatchReq.MessageID)
	if err != nil {
		log.Println("error while mark message readed: " + err.Error())
		return model.ErrMessageNotExist
	}

	iUserID, err := strconv.Atoi(userID)
	if err != nil {
		log.Println("error while mark message readed: can't parse user ID: ", err)
		return errors.New("can't parse user ID")
	}

	uintUserID := uint32(iUserID)

	err = CheckUserExist(uintUserID)
	if err != nil {
		log.Println("error while mark message readed: " + err.Error())
		return err
	}

	err = model.MarkMessageReadedUpdate(messagePatchReq.Readed, uintUserID, messagePatchReq.MessageID, messagePatchReq.ThreadID)
	if err != nil {
		log.Println("error while mark message readed: " + err.Error())
		return err
	}

	return nil
}

// GetCustomerMessagesList return slice with customers messages
func GetCustomerMessagesList(threadID, userID uint32, count, offset int) ([]*webpojo.MessageListResp, error) {
	if threadID == 0 {
		return nil, errors.New("error while get customer messages list: thread ID is 0")
	}

	if count == 0 {
		return nil, errors.New("error while get customer messages list: count is 0")
	}

	if offset < 0 {
		return nil, errors.New("error while get customer messages list: offset lessa then 0")
	}

	usersMessage, err := model.MessagesByUserID(threadID, userID, count, offset)
	if err != nil {
		log.Println("error while get customers messages list: " + err.Error())
		return nil, err
	}

	log.Println(usersMessage)

	var res []*webpojo.MessageListResp

	for _, v := range usersMessage {
		res = append(res, &webpojo.MessageListResp{
			ThreadID:   v.ThreadID,
			MessageID:  v.ID,
			FromUserID: v.FromUserID,
			FromFirst:  v.FromUserFirstName,
			FromLast:   v.FromUserLastName,
			ToUserID:   v.ToUserID,
			ToFirst:    v.ToUserFirstName,
			ToLast:     v.ToUserLastName,
			Content:    v.Content,
			Readed:     v.Readed,
			Created:    v.CreatedAt.String(),
		})
	}

	return res, nil
}

// GetCustomerMessageByID return single object customers messages
func GetCustomerMessageByID(userID, messageID string) (*webpojo.MessageListResp, error) {
	if userID == "" {
		return nil, errors.New("error while get customer message by ID: user ID is 0")
	}

	if messageID == "" {
		return nil, errors.New("error while get customer message by ID: message ID is empty")
	}

	usersMessage, err := model.MessageByID(userID, messageID)
	if err != nil {
		log.Println("error while get customers message by ID: " + err.Error())
		return nil, err
	}

	return &webpojo.MessageListResp{
		ThreadID:   usersMessage.ThreadID,
		MessageID:  usersMessage.ID,
		FromUserID: usersMessage.FromUserID,
		FromFirst:  usersMessage.FromUserFirstName,
		FromLast:   usersMessage.FromUserLastName,
		ToUserID:   usersMessage.ToUserID,
		ToFirst:    usersMessage.ToUserFirstName,
		ToLast:     usersMessage.ToUserLastName,
		Content:    usersMessage.Content,
		Readed:     usersMessage.Readed,
		Created:    usersMessage.CreatedAt.String(),
	}, nil
}

package controller

import (
	"app/constants"
	"app/model"
	"app/provider"
	"app/shared/session"
	"app/webpojo"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

// UserMessagePost send new message to user
func UserMessagePost(w http.ResponseWriter, r *http.Request) {
	sess := session.Instance(r)

	if sess == nil {
		log.Println("error while post new user's message: sess is nil")
		ReturnCodeError(w, errors.New("unauthorized"), http.StatusUnauthorized, constants.Msg_401)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("error while post new user's message: " + err.Error())
		ReturnCodeError(w, errors.New("internal server error"), http.StatusInternalServerError, constants.Msg_500)
		return
	}

	if len(body) == 0 {
		log.Println("error while post new user's message: empty json payoload")
		ReturnCodeError(w, errors.New("emtpy json payload"), http.StatusBadRequest, constants.Msg_400)
		return
	}

	log.Println("post new user's message: ", string(body))

	messagePostReq := &webpojo.MessagePostReq{}
	jsonErr := json.Unmarshal(body, &messagePostReq)
	if jsonErr != nil {
		log.Println("error while post new user's message: can't unmarshall request: " + jsonErr.Error())
		ReturnCodeError(w, errors.New("can't parse request"), http.StatusBadRequest, constants.Msg_400)
		return
	}

	if getUserID(sess) != fmt.Sprint(messagePostReq.FromUserID) {
		log.Println("error while post new user's message: user ID from session and in JSON request is not equal")
		ReturnCodeError(w, errors.New("wrong user ID"), http.StatusBadRequest, constants.Msg_400)
		return
	}

	resp, err := provider.PostNewMessage(getUserID(sess), messagePostReq)
	switch err {
	case nil:
		ReturnCodeJSONResponse(w, http.StatusOK, resp)
		return
	case model.ErrConstraintFails:
		log.Println("error while post new user's message: user or another agrs not found")
		ReturnCodeError(w, errors.New("can't create message: user or another agrs not found"), http.StatusNotFound, constants.Msg_404)
		return
	case model.ErrUserNotExist:
		log.Println("error while post new user's message: user not exist")
		ReturnCodeError(w, errors.New("can't create message: user not exist"), http.StatusNotFound, constants.Msg_404)
		return
	default:
		log.Println("error while post new user's message: " + err.Error())
		ReturnCodeError(w, errors.New("internal server error"), http.StatusInternalServerError, constants.Msg_500)
		return
	}
}

// MessageThreads returns list of messages threads
func MessageThreads(w http.ResponseWriter, r *http.Request) {
	// Get session
	sess := session.Instance(r)
	userID := getUserID(sess)

	body, readErr := ioutil.ReadAll(r.Body)
	if readErr != nil {
		log.Println("error while get customer messages threads: " + readErr.Error())
		ReturnCodeError(w, errors.New("can't read request body"), http.StatusInternalServerError, constants.Msg_500)
		return
	}

	if len(body) == 0 {
		sess.Save(r, w)
		log.Println("error while get customer messages threads: empty json payoload")
		ReturnCodeError(w, errors.New("emtpy json payload"), http.StatusBadRequest, constants.Msg_400)
		return
	}

	log.Println("get customer messages threads: ", string(body))
	messagesReq := webpojo.MessagesThreadsPostReq{}
	jsonErr := json.Unmarshal(body, &messagesReq)
	if jsonErr != nil {
		log.Println("error while get customer's messages threads: can't unmarshall request")
		ReturnCodeError(w, errors.New("can't parse request"), http.StatusBadRequest, constants.Msg_400)
		return
	}

	if messagesReq.Count < 1 {
		log.Println("error while get customer's message threads list: count < 1")
		ReturnCodeError(w, errors.New("bad_request: count less then 1"), http.StatusBadRequest, constants.Msg_400)
		return
	}

	if messagesReq.Offset < 0 {
		log.Println("error while get customer's message threads list: offset < 0")
		ReturnCodeError(w, errors.New("bad_request: offset less then 0"), http.StatusBadRequest, constants.Msg_400)
		return
	}

	list, err := provider.ThreadsByUserID(userID, messagesReq.Count, messagesReq.Offset)
	if err != nil {
		log.Println("error while get customer's messages threads: " + err.Error())
		ReturnCodeError(w, errors.New("internal server error"), http.StatusInternalServerError, constants.Msg_500)
		return
	}

	err = ReturnCodeJSONResp(w, list, http.StatusOK)
	if err != nil {
		log.Println("error while get customer's messages threads: error while return JSON response: " + err.Error())
		ReturnCodeError(w, errors.New("internal server error"), http.StatusInternalServerError, constants.Msg_500)
		return
	}
}

// MarkMessageAsReaded do mark on message
func MarkMessageAsReaded(w http.ResponseWriter, r *http.Request) {
	sess := session.Instance(r)

	if sess == nil {
		log.Println("error while patch message (set readed): sess is nil")
		ReturnCodeError(w, errors.New("unauthorized"), http.StatusUnauthorized, constants.Msg_401)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("error while patch message (set readed): " + err.Error())
		ReturnCodeError(w, errors.New("can't read request body"), http.StatusInternalServerError, constants.Msg_500)
		return
	}

	if len(body) == 0 {
		log.Println("error while patch customer message (set readed): empty json payoload")
		ReturnCodeError(w, errors.New("emtpy json payload"), http.StatusBadRequest, constants.Msg_400)
		return
	}

	log.Println("patch customer's message: ", string(body))

	messagePatchReq := &webpojo.UserMessagePatchReq{}
	jsonErr := json.Unmarshal(body, &messagePatchReq)
	if jsonErr != nil {
		log.Println("error while patch customer message (set readed): can't unmarshall request: " + jsonErr.Error())
		ReturnCodeError(w, errors.New("can't parse request"), http.StatusBadRequest, constants.Msg_400)
		return
	}

	log.Println("THREAD: ", messagePatchReq.ThreadID)
	err = provider.MarkMessageReaded(getUserID(sess), messagePatchReq)
	switch err {
	case nil:
		ReturnCodeError(w, errors.New(""), http.StatusOK, constants.Msg_200)
		return
	case model.ErrThreadNotExist:
		log.Println("error while patch customer message: thread not exist")
		ReturnCodeError(w, errors.New("can't patch message: thread not exist"), http.StatusNotFound, constants.Msg_404)
		return
	case model.ErrMessageNotExist:
		log.Println("error while patch customer message: message not exist")
		ReturnCodeError(w, errors.New("can't patch message: message not exist"), http.StatusNotFound, constants.Msg_404)
		return
	case model.ErrNoResult:
		log.Println("error while patch customer message (set readed): message not found")
		ReturnCodeError(w, errors.New("can't patch message: message not found"), http.StatusNotFound, constants.Msg_404)
		return
	case model.ErrUserNotExist:
		log.Println("error while patch customer message's message (set readed): user not exist")
		ReturnCodeError(w, errors.New("can't patch message: user not exist"), http.StatusNotFound, constants.Msg_404)
		return
	default:
		log.Println("error while patch customer message's message (set readed): " + err.Error())
		ReturnCodeError(w, errors.New("internal server error"), http.StatusInternalServerError, constants.Msg_500)
	}

}

// UserMessageList returns list with customer's messages
func UserMessageList(w http.ResponseWriter, r *http.Request) {
	// Get session
	sess := session.Instance(r)
	userID := getUserID(sess)

	iUserID, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		log.Println("error while get customer messages list: can't parse ID")
		ReturnCodeError(w, errors.New("internal server error"), http.StatusInternalServerError, constants.Msg_500)
		return
	}

	body, readErr := ioutil.ReadAll(r.Body)
	if readErr != nil {
		log.Println("error while get customer messages: " + readErr.Error())
		ReturnCodeError(w, errors.New("can't read request body"), http.StatusInternalServerError, constants.Msg_500)
		return
	}

	if len(body) == 0 {
		sess.Save(r, w)
		log.Println("error while get customer messages: empty json payoload")
		ReturnCodeError(w, errors.New("emtpy json payload"), http.StatusBadRequest, constants.Msg_400)
		return
	}

	log.Println("get customer message by ID: ", string(body))
	messagesReq := webpojo.MessagesListPostReq{}
	jsonErr := json.Unmarshal(body, &messagesReq)
	if jsonErr != nil {
		log.Println("error while get customer's message by ID: can't unmarshall request")
		ReturnCodeError(w, errors.New("can't parse request"), http.StatusBadRequest, constants.Msg_400)
		return
	}

	if messagesReq.Count < 1 {
		log.Println("error while get customer's message list: count < 1")
		ReturnCodeError(w, errors.New("bad_request: count less then 1"), http.StatusBadRequest, constants.Msg_400)
		return
	}

	if messagesReq.Offset < 0 {
		log.Println("error while get customer's message list: offset < 0")
		ReturnCodeError(w, errors.New("bad_request: offset less then 0"), http.StatusBadRequest, constants.Msg_400)
		return
	}

	list, err := provider.GetCustomerMessagesList(messagesReq.ThreadID, uint32(iUserID), messagesReq.Count, messagesReq.Offset)
	if err != nil {
		log.Println("error while get customer's messages list: " + err.Error())
		ReturnCodeError(w, errors.New("internal server error"), http.StatusInternalServerError, constants.Msg_500)
		return
	}

	err = ReturnCodeJSONResp(w, list, http.StatusOK)
	if err != nil {
		log.Println("error while get customer's messages list: error while return JSON response: " + err.Error())
		ReturnCodeError(w, errors.New("internal server error"), http.StatusInternalServerError, constants.Msg_500)
		return
	}
}

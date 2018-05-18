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
)

// AdminMessagePatch edit user's message
func AdminMessagePatch(w http.ResponseWriter, r *http.Request) {
	sess := session.Instance(r)

	if sess == nil {
		log.Println("error while patch user's message: sess is nil")
		ReturnCodeError(w, errors.New("unauthorized"), http.StatusUnauthorized, constants.Msg_401)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("error while patch user's message: " + err.Error())
		ReturnCodeError(w, errors.New("internal server error"), http.StatusInternalServerError, constants.Msg_500)
		return
	}

	if len(body) == 0 {
		log.Println("error while patch user's message: empty json payoload")
		ReturnCodeError(w, errors.New("emtpy json payload"), http.StatusBadRequest, constants.Msg_400)
		return
	}

	log.Println("patch new user's message: ", string(body))

	messagePatchReq := &webpojo.MessagePatchReq{}
	jsonErr := json.Unmarshal(body, &messagePatchReq)
	if jsonErr != nil {
		log.Println("error while patch user's message: can't unmarshall request: " + jsonErr.Error())
		ReturnCodeError(w, errors.New("can't parse request"), http.StatusBadRequest, constants.Msg_400)
		return
	}

	if getUserID(sess) != fmt.Sprint(messagePatchReq.FromUserID) {
		log.Println("error while patch user's message: user ID from session and in JSON request is not equal")
		ReturnCodeError(w, errors.New("wrong user ID"), http.StatusConflict, constants.Msg_409)
		return
	}

	err = provider.PatchMessage(messagePatchReq)
	switch err {
	case nil:
		ReturnCodeError(w, errors.New(""), http.StatusOK, constants.Msg_200)
		return
	case model.ErrThreadNotExist:
		log.Println("error while path message: thread not found")
		ReturnCodeError(w, errors.New("can't patch message: thread not found"), http.StatusNotFound, constants.Msg_404)
		return
	case model.ErrMessageNotExist:
		log.Println("error while path message: message not found")
		ReturnCodeError(w, errors.New("can't patch message: message not found"), http.StatusNotFound, constants.Msg_404)
		return
	case model.ErrUserNotExist:
		log.Println("error while path message: user not exist")
		ReturnCodeError(w, errors.New("can't patch message: user not exist"), http.StatusNotFound, constants.Msg_404)
		return
	default:
		log.Println("error while path message: " + err.Error())
		ReturnCodeError(w, errors.New("internal server error"), http.StatusInternalServerError, constants.Msg_500)
		return
	}
}

// AdminMessageDelete delete user's message
func AdminMessageDelete(w http.ResponseWriter, r *http.Request) {
	sess := session.Instance(r)

	if sess == nil {
		log.Println("error while delete customer's message: sess is nil")
		ReturnCodeError(w, errors.New("unauthorized"), http.StatusUnauthorized, constants.Msg_401)
		return
	}

	body, readErr := ioutil.ReadAll(r.Body)
	if readErr != nil {
		log.Println("error while delete customer's message: " + readErr.Error())
		ReturnError(w, readErr)
		return
	}

	if len(body) == 0 {
		sess.Save(r, w)
		log.Println("error while delete customer's message: empty json payoload")
		ReturnCodeError(w, errors.New("emtpy json payload"), http.StatusBadRequest, constants.Msg_400)
		return
	}

	log.Println("delete customer message: ", string(body))
	idReq := webpojo.MessageDeleteReq{}
	jsonErr := json.Unmarshal(body, &idReq)
	if jsonErr != nil {
		log.Println("error while delete customer's message: can't unmarshall request")
		ReturnCodeError(w, errors.New("can't parse request"), http.StatusBadRequest, constants.Msg_400)
		return
	}

	if fmt.Sprint(idReq.FromUserID) != getUserID(sess) {
		log.Println("error while delete customer's message: from user ID and user ID from session is not equal: ", idReq.FromUserID, " - ", getUserID(sess))
		ReturnCodeError(w, errors.New("wrong from user ID"), http.StatusConflict, constants.Msg_409)
		return
	}

	err := provider.DeleteMessage(idReq.FromUserID, idReq.MessageID, idReq.ThreadID)
	switch err {
	case nil:
		ReturnCodeError(w, errors.New(""), http.StatusOK, constants.Msg_200)
		return
	case model.ErrNoResult:
		ReturnCodeError(w, errors.New("error while delete customer's message: message not found"), http.StatusNotFound, constants.Msg_404)
		return
	case model.ErrUserNotExist:
		ReturnCodeError(w, errors.New("error while delete customer's message: user not exist"), http.StatusNotFound, constants.Msg_404)
		return
	case model.ErrThreadNotExist:
		ReturnCodeError(w, errors.New("error while delete customer's message: thread not exist"), http.StatusNotFound, constants.Msg_404)
		return
	case model.ErrMessageNotExist:
		ReturnCodeError(w, errors.New("error while delete customer's message: message not exist"), http.StatusNotFound, constants.Msg_404)
		return
	default:
		ReturnCodeError(w, errors.New("internal server error"), http.StatusInternalServerError, constants.Msg_500)
		return
	}
}

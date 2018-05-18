package controller

import (
	"app/constants"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/twinj/uuid"
)

const (
	//SessID - session id from the previous page
	SessID = "session_id"
	//SessRegisterAttempt for tracking register attempt
	SessRegisterAttempt = "register_attempt"
	//SessLoginAttempt - Name of the session variable that tracks login attempts
	SessLoginAttempt = "login_attempt"
	//SessDbAttempt tracking db attempts
	SessDbAttempt = "db_attempt"
	SessLeadID    = "lead_id"
	UserID        = "user_id"
	UserName      = "username"
	UserRole      = "user_role"
)

//RecordLoginAttempt increments the number of login attempts in sessions variable
func RecordLoginAttempt(sess *sessions.Session) {
	// Log the attempt
	if sess.Values[SessLoginAttempt] == nil {
		sess.Values[SessLoginAttempt] = 0
	} else {
		sess.Values[SessLoginAttempt] = sess.Values[SessLoginAttempt].(int) + 1
	}
}

//SessRegisterAttempt increments the number of login attempts in sessions variable
func RecordRegisterAttempt(sess *sessions.Session) {
	// Log the attempt
	if sess.Values[SessRegisterAttempt] == nil {
		sess.Values[SessRegisterAttempt] = 0
	} else {
		sess.Values[SessRegisterAttempt] = sess.Values[SessRegisterAttempt].(int) + 1
	}
}

//RecordDbAttempt increments the number of DB attempts in sessions variable
func RecordDbAttempt(sess *sessions.Session) {
	// Log the db attempt
	if sess.Values[SessDbAttempt] == nil {
		sess.Values[SessDbAttempt] = 1
	} else {
		sess.Values[SessDbAttempt] = sess.Values[SessDbAttempt].(int) + 1
	}
}

//RecordSessID saving the loan id into session
func RecordSessID(sess *sessions.Session) {
	sess.Values[SessID] = uuid.NewV4().String()
}

//RecordLeadID saving the lead id into session
func RecordLeadID(sess *sessions.Session, leadId string) {
	sess.Values[SessLeadID] = leadId
}

//APIReturnError for debugging purpose. Might change later
func APIReturnError(w http.ResponseWriter, err error) {
	fmt.Fprintf(w, "{\"statusCode\": 500, \"message\": \"Internal Server Error\", \"error\": \"%s\"}", err)
}

//APIReturnErrorWithCustomMsg for custom response
func APIReturnErrorWithCustomMsg(w http.ResponseWriter, statusCode uint16, msg string, err error) {
	fmt.Fprint(w, "{\"statusCode\":"+fmt.Sprint(statusCode)+", \"message\": \""+msg+"\", \"error\": \"%v\"}", err)
}

//APIReturnErrorWithCustomMsg for custom response
func GetCustomErrorMsg(statusCode uint16, msg string, err error) string {
	return "{\"statusCode\":" + fmt.Sprint(statusCode) + ", \"message\": \"" + msg + "\", \"error\": \"" + err.Error() + "\"}"
}

//ReturnError return error message to clients
func ReturnError(w http.ResponseWriter, err error) {
	fmt.Fprint(w, fmt.Sprintf("{\"error\": \"%v\"}", err))
}

//ReturnCodeError return error message to clients with current HTTP error code
func ReturnCodeError(w http.ResponseWriter, err error, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\"statusCode\":" + fmt.Sprint(code) + ", \"message\": \"" + msg + "\", \"error\": \"" + err.Error() + "\"}"))
}

// ReturnCodeJSONResponse return marshaled json response
func ReturnCodeJSONResponse(w http.ResponseWriter, code int, resp interface{}) {
	bs, err := json.Marshal(resp)
	if err != nil {
		ReturnCodeError(w, err, http.StatusInternalServerError, constants.Msg_500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(bs)
}

func ReturnJsonResp(w http.ResponseWriter, resp interface{}) {
	bs, err := json.Marshal(resp)
	if err != nil {
		log.Println("error while marshal JSON response: " + err.Error())
		ReturnCodeError(w, errors.New("internal server error"), http.StatusInternalServerError, constants.Msg_500)

	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(bs)
	if err != nil {
		log.Println("error while write JSON response: " + err.Error())
		ReturnCodeError(w, errors.New("internal server error"), http.StatusInternalServerError, constants.Msg_500)
	}
}

func ReturnCodeJSONResp(w http.ResponseWriter, resp interface{}, code int) error {
	bs, err := json.Marshal(resp)
	if err != nil {
		log.Println("error while marshal JSON response: " + err.Error())
		ReturnCodeError(w, errors.New("internal server error"), http.StatusInternalServerError, constants.Msg_500)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(bs)
	if err != nil {
		log.Println("error while write JSON response: " + err.Error())
		ReturnCodeError(w, errors.New("internal server error"), http.StatusInternalServerError, constants.Msg_500)
		return err
	}

	return nil
}

func ReturnNoEscapeCodeJSONResp(w http.ResponseWriter, resp interface{}, code int) error {
	bs, err := MarshalNoExpape(resp)
	if err != nil {
		log.Println("error while marshal JSON response: " + err.Error())
		ReturnCodeError(w, errors.New("internal server error"), http.StatusInternalServerError, constants.Msg_500)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(bs)
	if err != nil {
		log.Println("error while write JSON response: " + err.Error())
		ReturnCodeError(w, errors.New("internal server error"), http.StatusInternalServerError, constants.Msg_500)
		return err
	}

	return nil
}

func MarshalNoExpape(data interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(data)
	return buffer.Bytes(), err
}

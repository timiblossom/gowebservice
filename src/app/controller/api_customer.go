package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"app/constants"
	"app/model"
	"app/provider"
	fas "app/shared/files_id_storage"
	"app/shared/passhash"
	"app/shared/session"
	"app/webpojo"
)

const (
	uploadFileFormName = "uploadfile"
)

// CustomerRegisterPost handles customers registration JSON submission
func CustomerRegisterPost(w http.ResponseWriter, r *http.Request) {
	// Get session
	sess := session.Instance(r)
	RecordRegisterAttempt(sess)

	// Prevent brute force login attempts by not hitting MySQL and pretending like it was invalid :-)
	if sess.Values["register_attempt"] != nil && sess.Values["register_attempt"].(int) >= 10 {
		log.Println("error while registen new customer: brute force register prevented: count=", sess.Values["register_attempt"].(int))
		ReturnCodeError(w, errors.New("too many requests"), http.StatusTooManyRequests, constants.Msg_429)
		return
	}

	body, readErr := ioutil.ReadAll(r.Body)
	if readErr != nil {
		log.Println("error while read register customer body: " + readErr.Error())
		ReturnCodeError(w, errors.New("can't read request body"), http.StatusInternalServerError, constants.Msg_500)
		return
	}

	if len(body) == 0 {
		sess.Save(r, w)
		log.Println("error while register new customer: empty json payoload")
		ReturnCodeError(w, errors.New("request body is nil"), http.StatusBadRequest, constants.Msg_400)
		return
	}

	log.Println("register new customer info: ", string(body))
	regReq := webpojo.UserCreateReq{}
	jsonErr := json.Unmarshal(body, &regReq)
	if jsonErr != nil {
		log.Println("error while register new customer: can't unmarshall request")
		ReturnCodeError(w, errors.New("can't parse request"), http.StatusBadRequest, constants.Msg_400)
		return
	}

	err := trimUserRegisterInfo(&regReq)
	if err != nil {
		log.Println("error while register new customer: can't trim spaces")
		ReturnCodeError(w, errors.New("can't parse request body"), http.StatusInternalServerError, constants.Msg_500)
		return
	}

	log.Println(fmt.Sprintf("register new cusromer with data: email=%s, first=%s, last=%s, pass=%s", regReq.Email, regReq.FirstName, regReq.LastName, regReq.Password))

	// Validate with required fields
	validate, message := validateRegisterInfo(r, &regReq, constants.CustomerRole)
	if !validate {
		log.Println("error while register new customer: invalid reg request - " + message)
		sess.Save(r, w)
		ReturnCodeError(w, errors.New(message), http.StatusBadRequest, constants.Msg_400)
		return
	}

	password, errp := passhash.HashString(regReq.Password)
	// If password hashing failed
	if errp != nil {
		log.Println("error while register new customer: password hashing failed")
		sess.Save(r, w)
		ReturnCodeError(w, errors.New("can't hash password"), http.StatusInternalServerError, constants.Msg_500)
		return
	}

	isExist, err := provider.IsUserExist(regReq.Email)
	if err != nil {
		log.Println("error while register new customer/check is exist: " + err.Error())
		ReturnCodeError(w, errors.New("internal error"), http.StatusInternalServerError, constants.Msg_500)
		return
	}

	switch isExist {
	case false:
		err := provider.RegisterNewCustomer(&regReq, password)
		if err != nil {
			log.Println("error while creating new customer: " + err.Error())
			sess.Save(r, w)
			// better way is no return db errors to front-end - this info can be use for hack
			ReturnCodeError(w, errors.New("internal error"), http.StatusInternalServerError, constants.Msg_500)
			return
		}

		log.Println("new account created successfully for: " + regReq.Email)
		ReturnCodeError(w, errors.New(""), http.StatusOK, constants.Msg_200)
		sess.Save(r, w)
	case true:
		log.Println("error while register new customer: user already exist")
		sess.Save(r, w)
		ReturnCodeError(w, errors.New("user already exist"), http.StatusConflict, constants.Msg_409)
	}
}

// CustomerLoginPost for handling customer login's post request
func CustomerLoginPost(w http.ResponseWriter, r *http.Request) {
	sess := session.Instance(r)
	RecordRegisterAttempt(sess)

	// Prevent brute force login attempts by not hitting MySQL and pretending like it was invalid :-)
	if sess.Values[SessLoginAttempt] != nil && sess.Values[SessLoginAttempt].(int) >= 10 {
		log.Println("error while login customer: brute force login prevented: count=", sess.Values["register_attempt"].(int))
		ReturnCodeError(w, errors.New("too many requests"), http.StatusTooManyRequests, constants.Msg_429)
		return
	}

	body, readErr := ioutil.ReadAll(r.Body)
	if readErr != nil {
		log.Println("error while login customer: can't read request body: " + readErr.Error())
		ReturnCodeError(w, errors.New("can't read request body"), http.StatusInternalServerError, constants.Msg_500)
		return
	}

	if len(body) == 0 {
		sess.Save(r, w)
		log.Println("error while login customer: empty json payoload")
		ReturnCodeError(w, errors.New("request body is nil"), http.StatusBadRequest, constants.Msg_400)
		return
	}

	log.Println("login customer info: ", string(body))
	loginReq := webpojo.UserLoginReq{}
	jsonErr := json.Unmarshal(body, &loginReq)
	if jsonErr != nil {
		log.Println("error while login customer: can't unmarshall request")
		ReturnCodeError(w, errors.New("can't parse request"), http.StatusBadRequest, constants.Msg_400)
		return
	}

	//should check for expiration
	if sess.Values[UserID] != nil && sess.Values[UserName] == loginReq.Username {
		log.Println("WARNING: twice user login: user with this credentials already exist")
		ReturnCodeError(w, errors.New("user already logged in"), http.StatusConflict, constants.Msg_409)
		return
	}

	user, err := provider.GetUserByEmail(loginReq.Username)
	switch err {
	case model.ErrNoResult:
		log.Println("error while login customer: no result, attempts=", sess.Values[SessLoginAttempt])
		sess.Save(r, w)
		ReturnCodeError(w, errors.New("user not found"), http.StatusNotFound, constants.Msg_404)
		return

	case nil:
		if passhash.MatchString(user.Password, loginReq.Password) {
			log.Println("new customer successfully logged in: " + loginReq.Username)
			session.Empty(sess)
			log.Println(*user)
			sess.Values[UserID] = user.UserID()
			sess.Values[UserName] = loginReq.Username
			sess.Values[UserRole] = user.UserRole
			sess.Save(r, w)
			ReturnJsonResp(w, makeCustomerLoginResp(constants.StatusCode_200, constants.Msg_200, "/api/customer/login", user))
			return
		}

		ReturnCodeError(w, errors.New("wrong password"), http.StatusNotFound, constants.Msg_404)
		return

	default:
		log.Println("error while login customer: " + err.Error())
		ReturnCodeError(w, errors.New("internal error"), http.StatusInternalServerError, constants.Msg_500)
		return
	}
}

// CustomerLogoutPost clears the session and logs the user out
func CustomerLogoutPost(w http.ResponseWriter, r *http.Request) {
	// Get session
	sess := session.Instance(r)

	if sess == nil {
		log.Println("error while logout customer: sess is nil")
		ReturnCodeError(w, errors.New("session is nil"), http.StatusUnauthorized, constants.Msg_401)
		return
	}

	if sess.Values[UserID] == nil {
		log.Println("error while logout customer: sess id is nil")
		ReturnCodeError(w, errors.New("unauthorized"), http.StatusUnauthorized, constants.Msg_401)
		return
	}

	// If user is authenticated
	log.Println("logout customer with userID=", sess.Values[UserID], " userName=", sess.Values[UserName])
	session.Empty(sess)
	sess.Save(r, w)
	ReturnCodeError(w, errors.New(""), http.StatusOK, constants.Msg_200)
}

// CustomerProfileGetInfo return user's profile info
func CustomerProfileGetInfo(w http.ResponseWriter, r *http.Request) {
	// Get session
	sess := session.Instance(r)

	if sess == nil {
		log.Println("error while getting customer profile info: session is nil")
		ReturnCodeError(w, errors.New("internal server error"), http.StatusInternalServerError, constants.Msg_500)
		return
	}

	body, readErr := ioutil.ReadAll(r.Body)
	if readErr != nil {
		log.Println("error while get profile info: can't read request body: " + readErr.Error())
		ReturnCodeError(w, errors.New("can't read body"), http.StatusInternalServerError, constants.Msg_500)
		return
	}

	if len(body) == 0 {
		sess.Save(r, w)
		log.Println("error while get profile info: empty json payoload")
		ReturnCodeError(w, errors.New("emtpy json payload"), http.StatusBadRequest, constants.Msg_400)
		return
	}

	log.Println("profile get info request: ", string(body))
	idReq := webpojo.IDRequest{}
	jsonErr := json.Unmarshal(body, &idReq)
	if jsonErr != nil {
		log.Println("error while get customer profile info: can't unmarshall request")
		ReturnCodeError(w, errors.New("internal server error: can't unmarshall json"), http.StatusInternalServerError, constants.Msg_500)
		return
	}

	customerInfoRepsonse, err := provider.GetCustomerProfile(fmt.Sprint(idReq.ID))
	switch err {
	case model.ErrNoResult:
		log.Println("customer profile with id=", idReq.ID, " has no info")
		ReturnCodeError(w, errors.New("profile not foubnd"), http.StatusNotFound, constants.Msg_404)
		return
	case nil:
		ReturnCodeJSONResponse(w, http.StatusOK, customerInfoRepsonse)
	default:
		ReturnCodeError(w, errors.New("internal server error"), http.StatusInternalServerError, constants.Msg_500)
	}
}

// CustomerProfilePatch patch customer profile
func CustomerProfilePatch(w http.ResponseWriter, r *http.Request) {
	// Get session
	sess := session.Instance(r)

	if sess == nil {
		log.Println("error while patch customer profile info: sess is nil")
		ReturnCodeError(w, errors.New("session is nil"), http.StatusUnauthorized, constants.Msg_401)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("error while patch customer profile info: " + err.Error())
		ReturnCodeError(w, errors.New("can't read body"), http.StatusInternalServerError, constants.Msg_500)
		return
	}

	if len(body) == 0 {
		sess.Save(r, w)
		log.Println("error while patch customer profile info: empty json payoload")
		ReturnCodeError(w, errors.New("emtpy json payload"), http.StatusBadRequest, constants.Msg_400)
		return
	}

	log.Println("patch customer info: ", string(body))
	customerPatchReq := webpojo.CustomerInfoPojoPatchReq{}
	jsonErr := json.Unmarshal(body, &customerPatchReq)
	if jsonErr != nil {
		log.Println("error while patch customer info: can't unmarshall request")
		ReturnCodeError(w, errors.New("can't parse request"), http.StatusBadRequest, constants.Msg_400)
		return
	}

	// Validate with required fields
	validate, message := validatePassword(customerPatchReq.Password)
	if !validate {
		log.Println("error while patch customer: invalid password: " + message)
		sess.Save(r, w)
		ReturnCodeError(w, errors.New(message), http.StatusBadRequest, constants.Msg_400)
		return
	}

	customerPatchReq.Password, err = passhash.HashString(customerPatchReq.Password)
	// If password hashing failed
	if err != nil {
		log.Println("error while patch new customer: password hashing failed")
		ReturnCodeError(w, errors.New("can't hash password"), http.StatusInternalServerError, constants.Msg_500)
		return
	}

	userExist, err := provider.CheckIsUserExistString(customerPatchReq.UserID)
	if err != nil {
		log.Println("error while patch customer profile info: " + err.Error())
		ReturnCodeError(w, errors.New("internal server error"), http.StatusInternalServerError, constants.Msg_500)
		return
	}

	if !userExist {
		log.Println("error while patch customer profile info: user not exist")
		ReturnCodeError(w, errors.New("user not exist"), http.StatusNotFound, constants.Msg_404)
		return
	}

	err = provider.PatchCustomerProfileInfo(&customerPatchReq)
	if err != nil {
		log.Println("error while patch customer profile info: " + err.Error())
		ReturnCodeError(w, errors.New("internal server error"), http.StatusInternalServerError, constants.Msg_500)
		return
	}

	ReturnCodeError(w, errors.New(""), http.StatusOK, constants.Msg_200)
}

// CustomerUploadFilePost handler handle all uploads request
func CustomerUploadFilePost(w http.ResponseWriter, r *http.Request) {
	// Get session
	sess := session.Instance(r)

	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile(uploadFileFormName)
	if err != nil {
		log.Println("error while parse form file name: " + err.Error())
		ReturnCodeError(w, errors.New("can't parse form file name: "+err.Error()), http.StatusBadRequest, constants.Msg_400)
		return
	}
	escapedFileName := strings.Replace(url.QueryEscape(handler.Filename), "+", "%20", -1)

	lastID, err := provider.SaveFileDataInDB(escapedFileName, "raw", getUserID(sess))
	if err != nil {
		log.Println("error while add new file info to DB: " + err.Error())
		ReturnCodeError(w, errors.New("internal server error"), http.StatusInternalServerError, constants.Msg_500)
		return
	}

	userName := getUserName(sess)
	err = provider.UploadFile(&file, handler, fmt.Sprint(lastID)+filepath.Ext(handler.Filename), userName)
	if err != nil {
		log.Println("error while upload new users file: " + err.Error())
		ReturnCodeError(w, errors.New("internal server error"), http.StatusInternalServerError, constants.Msg_500)
		return
	}

	fakeFileID := fas.GetNewFileID(int(lastID), sess.ID)
	ReturnCodeJSONResponse(w, http.StatusOK, webpojo.FileIDResponse{FileID: fakeFileID})
}

// CustomerFilesListGet handler handle all uploads request
func CustomerFilesListGet(w http.ResponseWriter, r *http.Request) {
	// Get session
	sess := session.Instance(r)
	userID := getUserID(sess)

	list, err := provider.GetFileList(userID, sess.ID)
	if err != nil {
		log.Println("error while get user files list: " + err.Error())
		ReturnCodeError(w, errors.New("internal server error"), http.StatusInternalServerError, constants.Msg_500)
		return
	}

	err = ReturnNoEscapeCodeJSONResp(w, list, http.StatusOK)
	if err != nil {
		log.Println("error while return JSON response: " + err.Error())
		ReturnCodeError(w, errors.New("internal server error"), http.StatusInternalServerError, constants.Msg_500)
		return
	}
}

// CustomerFileDelete not implemented yet
func CustomerFileDelete(w http.ResponseWriter, r *http.Request) {
	sess := session.Instance(r)

	if r.Body == nil {
		log.Println("error while delete customer file: request body is nil")
		ReturnCodeError(w, errors.New("can't read request body"), http.StatusInternalServerError, constants.Msg_500)
		return
	}

	body, readErr := ioutil.ReadAll(r.Body)
	if readErr != nil {
		log.Println("error while delete customer file: " + readErr.Error())
		ReturnCodeError(w, errors.New("can't read request body"), http.StatusInternalServerError, constants.Msg_500)
		return
	}

	if len(body) == 0 {
		log.Println("error while delete customer: empty json payoload")
		ReturnCodeError(w, errors.New("emtpy json payload"), http.StatusBadRequest, constants.Msg_400)
		return
	}

	log.Println("customer file delete request: ", string(body))
	idReq := webpojo.IDRequestString{}
	jsonErr := json.Unmarshal(body, &idReq)
	if jsonErr != nil {
		log.Println("error while delete customer file: can't unmarshall request")
		ReturnCodeError(w, errors.New("can't parse request"), http.StatusBadRequest, constants.Msg_400)
		return
	}

	realFileID, err := fas.GetRealID(idReq.ID, sess.ID)
	if err != nil {
		log.Println("error while get real file ID: id not found")
		ReturnCodeError(w, errors.New("bad file id"), http.StatusBadRequest, constants.Msg_400)
		return
	}

	err = provider.FileDelete(getUserID(sess), realFileID)
	switch err {
	case model.ErrNoResult:
		ReturnCodeError(w, errors.New("not_found"), http.StatusNotFound, constants.Msg_404)
		return
	case nil:
		ReturnCodeError(w, errors.New(""), http.StatusOK, constants.Msg_200)
		return
	default:
		log.Println("error while delete file from database: " + err.Error())
		ReturnCodeError(w, errors.New("internal server error"), http.StatusInternalServerError, constants.Msg_500)

	}
}

// CustomerServeFileGet return one file info
func CustomerServeFileGet(w http.ResponseWriter, r *http.Request) {
	sess := session.Instance(r)

	if r.Body == nil {
		log.Println("error while get customer file info: body is nil")
		ReturnCodeError(w, errors.New("body is nil"), http.StatusBadRequest, constants.Msg_400)
		return
	}

	body, readErr := ioutil.ReadAll(r.Body)
	if readErr != nil {
		log.Println("error while get customer file info: can't read request body")
		ReturnCodeError(w, errors.New("can't read request body"), http.StatusInternalServerError, constants.Msg_500)
		return
	}

	if len(body) == 0 {
		sess.Save(r, w)
		log.Println("error while get customer file info: empty json payoload")
		ReturnCodeError(w, errors.New("emtpy json payload"), http.StatusBadRequest, constants.Msg_400)
		return
	}

	log.Println("customer get file info request: ", string(body))
	idReq := webpojo.IDRequestString{}
	jsonErr := json.Unmarshal(body, &idReq)
	if jsonErr != nil {
		log.Println("error while get customer file info: can't unmarshall request")
		ReturnCodeError(w, errors.New("can't parse request"), http.StatusBadRequest, constants.Msg_400)
		return
	}

	realFileID, err := fas.GetRealID(idReq.ID, sess.ID)
	if err != nil {
		log.Println("error while get customer file info: error while get real file ID: id not found")
		ReturnCodeError(w, errors.New("bad file id"), http.StatusBadRequest, constants.Msg_400)
		return
	}

	userFile, err := provider.GetFile(getUserID(sess), fmt.Sprint(realFileID), sess.ID)
	if err != nil {
		log.Println("error while get customer file info: error while get user file: " + err.Error())
		ReturnCodeError(w, errors.New("internal server error"), http.StatusInternalServerError, constants.Msg_500)
		return
	}

	ReturnNoEscapeCodeJSONResp(w, userFile, http.StatusOK)
}

// // CustomerMessageByID returns single customer message
// func CustomerMessageByID(w http.ResponseWriter, r *http.Request) {
// 	sess := session.Instance(r)

// 	if sess == nil {
// 		log.Println("error while get customer's message by ID: sess is nil")
// 		ReturnCodeError(w, errors.New("unauthorized"), http.StatusUnauthorized, constants.Msg_401)
// 		return
// 	}

// 	body, readErr := ioutil.ReadAll(r.Body)
// 	if readErr != nil {
// 		log.Println("error while get customer messages: " + readErr.Error())
// 		ReturnCodeError(w, errors.New("can't read request body"), http.StatusInternalServerError, constants.Msg_500)
// 		return
// 	}

// 	if len(body) == 0 {
// 		sess.Save(r, w)
// 		log.Println("error while get customer messages: empty json payoload")
// 		ReturnCodeError(w, errors.New("emtpy json payload"), http.StatusBadRequest, constants.Msg_400)
// 		return
// 	}

// 	log.Println("get customer message by ID: ", string(body))
// 	idReq := webpojo.IDRequest{}
// 	jsonErr := json.Unmarshal(body, &idReq)
// 	if jsonErr != nil {
// 		log.Println("error while get customer's message by ID: can't unmarshall request")
// 		ReturnCodeError(w, errors.New("can't parse request"), http.StatusBadRequest, constants.Msg_400)
// 		return
// 	}

// 	message, err := provider.GetCustomerMessageByID(getUserID(sess), fmt.Sprint(idReq.ID))
// 	switch err {
// 	case nil:
// 		ReturnCodeJSONResp(w, message, http.StatusOK)
// 		return
// 	case model.ErrNoResult:
// 		log.Println("error while get customer's message by ID: not found")
// 		ReturnCodeError(w, errors.New("can't get message: message not found"), http.StatusNotFound, constants.Msg_404)
// 		return
// 	default:
// 		log.Println("error while get customer's message by ID: " + err.Error())
// 		ReturnCodeError(w, errors.New("internal server error"), http.StatusInternalServerError, constants.Msg_500)
// 		return
// 	}
// }

func makeCustomerLoginResp(statusCode uint16, msg string, url string, user *model.User) *webpojo.UserLoginResp {
	if user == nil {
		log.Println("error while make customer login response: user struct is nil")
		return nil
	}

	customerLoginResp := &webpojo.UserLoginResp{}
	customerLoginResp.StatusCode = statusCode
	customerLoginResp.Message = msg
	customerLoginResp.URL = url
	customerLoginResp.Email = user.Email
	customerLoginResp.FirstName = user.FirstName
	customerLoginResp.LastName = user.LastName
	customerLoginResp.UserRole = constants.CustomerRole

	return customerLoginResp
}

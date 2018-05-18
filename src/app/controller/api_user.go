package controller

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"app/constants"
	"app/model"
	"app/shared/passhash"
	"app/shared/session"
	"app/webpojo"
)

// UserRegisterPost handles the registration JSON submission
func UserRegisterPost(w http.ResponseWriter, r *http.Request) {
	// Get session
	sess := session.Instance(r)

	// Prevent brute force login attempts by not hitting MySQL and pretending like it was invalid :-)
	if sess.Values["register_attempt"] != nil && sess.Values["register_attempt"].(int) >= 5 {
		log.Println("Brute force register prevented")
		http.Redirect(w, r, "/not_found", http.StatusFound)
		return
	}

	body, readErr := ioutil.ReadAll(r.Body)
	if readErr != nil {
		log.Println(readErr)
		ReturnError(w, readErr)
		return
	}

	var regResp webpojo.UserCreateResp
	if len(body) == 0 {
		log.Println("Empty json payload")
		RecordRegisterAttempt(sess)
		sess.Save(r, w)
		regResp = webpojo.UserCreateResp{constants.StatusCode_400, constants.Msg_400}
		bs, err := json.Marshal(regResp)
		if err != nil {
			ReturnError(w, err)
			return
		}
		fmt.Fprint(w, string(bs))
		return
	}

	//log.Println("r.Body", string(body))
	regReq := webpojo.UserCreateReq{}
	jsonErr := json.Unmarshal(body, &regReq)
	if jsonErr != nil {
		log.Println(jsonErr)
		ReturnError(w, jsonErr)
		return
	}
	log.Println(regReq.Email)

	// Validate with required fields
	if validate, _ := validateRegisterInfo(r, &regReq, constants.DefaultRole); !validate {
		log.Println("Invalid reg request! Missing field")
		RecordRegisterAttempt(sess)
		sess.Save(r, w)
		regResp = webpojo.UserCreateResp{constants.StatusCode_400, constants.Msg_400}
		bs, err := json.Marshal(regResp)
		if err != nil {
			ReturnError(w, err)
			return
		}
		fmt.Fprint(w, string(bs))
		return
	}

	password, errp := passhash.HashString(regReq.Password)

	// If password hashing failed
	if errp != nil {
		log.Println(errp)
		RecordRegisterAttempt(sess)
		sess.Save(r, w)
		regResp = webpojo.UserCreateResp{constants.StatusCode_500, constants.Msg_500}
		bs, err := json.Marshal(regResp)
		if err != nil {
			ReturnError(w, err)
			return
		}
		fmt.Fprint(w, string(bs))
		return
	}

	// Get database result
	_, err := model.UserByEmail(regReq.Email)

	if err == model.ErrNoResult { // If success (no user exists with that email)
		ex := model.UserCreate(regReq.FirstName, regReq.LastName, regReq.Email, password)
		// Will only error if there is a problem with the query
		if ex != nil {
			log.Println(ex)
			RecordRegisterAttempt(sess)
			sess.Save(r, w)
			regResp = webpojo.UserCreateResp{constants.StatusCode_500, constants.Msg_500}
			bs, err := json.Marshal(regResp)
			if err != nil {
				ReturnError(w, err)
				return
			}
			fmt.Fprint(w, string(bs))
		} else {
			log.Println("Account created successfully for: " + regReq.Email)
			RecordRegisterAttempt(sess)
			sess.Save(r, w)
			regResp = webpojo.UserCreateResp{constants.StatusCode_200, constants.Msg_200}
			bs, err := json.Marshal(regResp)
			if err != nil {
				ReturnError(w, err)
				return
			}
			fmt.Fprint(w, string(bs))
		}
	} else if err != nil { // Catch all other errors
		log.Println(err)
		RecordRegisterAttempt(sess)
		sess.Save(r, w)
		regResp = webpojo.UserCreateResp{constants.StatusCode_500, constants.Msg_500}
		bs, err := json.Marshal(regResp)
		if err != nil {
			ReturnError(w, err)
			return
		}
		fmt.Fprint(w, string(bs))
	} else { // Else the user already exists
		log.Println("User already existed!!!")
		RecordRegisterAttempt(sess)
		sess.Save(r, w)
		regResp = webpojo.UserCreateResp{constants.StatusCode_400, constants.Msg_400}
		bs, err := json.Marshal(regResp)
		if err != nil {
			ReturnError(w, err)
			return
		}
		fmt.Fprint(w, string(bs))
	}
}

//UserLoginPost for handling admin login's post request
func UserLoginPost(w http.ResponseWriter, r *http.Request) {
	sess := session.Instance(r)
	var loginResp webpojo.UserLoginResp

	// Prevent brute force login attempts by not hitting MySQL and pretending like it was invalid :-)
	if sess.Values[SessLoginAttempt] != nil && sess.Values[SessLoginAttempt].(int) >= 5 {
		log.Println("Brute force login prevented")
		loginResp = makeUserLoginResp(constants.StatusCode_429, constants.Msg_429, "/api/admin/login")
		ReturnJsonResp(w, loginResp)
		return
	}

	body, readErr := ioutil.ReadAll(r.Body)
	if readErr != nil {
		log.Println(readErr)
		ReturnError(w, readErr)
		return
	}

	if len(body) == 0 {
		log.Println("Empty json payload")
		RecordLoginAttempt(sess)
		sess.Save(r, w)
		loginResp = makeUserLoginResp(constants.StatusCode_400, constants.Msg_400, "/api/admin/login")
		ReturnJsonResp(w, loginResp)
		return
	}

	//log.Println("r.Body", string(body))
	loginReq := webpojo.UserLoginReq{}
	jsonErr := json.Unmarshal(body, &loginReq)
	if jsonErr != nil {
		log.Println(jsonErr)
		ReturnError(w, jsonErr)
		return
	}
	log.Println(loginReq.Username)

	//should check for expiration
	if sess.Values[UserID] != nil && sess.Values[UserName] == loginReq.Username {
		log.Println("Already signed in - session is valid!!")
		sess.Save(r, w) //Should also start a new expiration
		loginResp = makeUserLoginResp(constants.StatusCode_200, constants.Msg_200, "/api/admin/leads")
		ReturnJsonResp(w, loginResp)
		return
	}

	result, dbErr := model.UserByEmail(loginReq.Username)
	if dbErr == model.ErrNoResult {
		log.Println("Login attempt: ", sess.Values[SessLoginAttempt])
		RecordLoginAttempt(sess)
		sess.Save(r, w)
		loginResp = makeUserLoginResp(constants.StatusCode_204, constants.Msg_204, "/api/admin/login")
	} else if dbErr != nil {
		log.Println(dbErr)
		RecordLoginAttempt(sess)
		sess.Save(r, w)
		loginResp = makeUserLoginResp(constants.StatusCode_500, constants.Msg_500, "/error")
	} else if passhash.MatchString(result.Password, loginReq.Password) {
		log.Println("Login successfully")
		session.Empty(sess)
		sess.Values[UserID] = result.UserID()
		sess.Values[UserName] = loginReq.Username
		sess.Values[UserRole] = result.UserRole
		sess.Save(r, w) //Should also store expiration
		loginResp = webpojo.UserLoginResp{}
		loginResp.StatusCode = constants.StatusCode_200
		loginResp.Message = constants.Msg_200
		loginResp.URL = "/api/admin/leads"
		loginResp.FirstName = result.FirstName
		loginResp.LastName = result.LastName
		loginResp.UserRole = result.UserRole
		loginResp.Email = loginReq.Username
	} else {
		log.Println("Login attempt: ", sess.Values[SessLoginAttempt])
		RecordLoginAttempt(sess)
		sess.Save(r, w)
		loginResp = makeUserLoginResp(constants.StatusCode_404, constants.Msg_404, "/api/admin/login")
	}

	ReturnJsonResp(w, loginResp)
}

//UserUpdatePost for handling user update post request
func UserUpdatePost(w http.ResponseWriter, r *http.Request) {
	sess := session.Instance(r)
	var updateReq webpojo.UserUpdateReq
	var updateResp = webpojo.UserUpdateResp{}

	// Prevent brute force login attempts by not hitting MySQL and pretending like it was invalid :-)
	if sess.Values[SessLoginAttempt] == nil ||
		(sess.Values[UserRole] != webpojo.UserSupervisor && sess.Values[UserRole] != webpojo.UserAdmin) {
		log.Println("Authorized request")
		updateResp.StatusCode = constants.StatusCode_429
		updateResp.Message = constants.Msg_429

		ReturnJsonResp(w, updateResp)
		return
	}

	body, readErr := ioutil.ReadAll(r.Body)
	if readErr != nil {
		log.Println(readErr)
		ReturnError(w, readErr)
		return
	}

	if len(body) == 0 {
		log.Println("Empty json payload")
		RecordLoginAttempt(sess)
		sess.Save(r, w)
		updateResp.StatusCode = constants.StatusCode_400
		updateResp.Message = constants.Msg_400
		ReturnJsonResp(w, updateResp)
		return
	}

	//log.Println("r.Body", string(body))
	updateReq = webpojo.UserUpdateReq{}
	jsonErr := json.Unmarshal(body, &updateReq)
	if jsonErr != nil {
		log.Println(jsonErr)
		ReturnError(w, jsonErr)
		return
	}
	log.Println(fmt.Sprintf("%v is updating user: %v", sess.Values[UserName], updateReq.Email))

	user := model.User{}
	user.Email = updateReq.Email
	user.Password, _ = passhash.HashString(updateReq.Password)
	user.FirstName = updateReq.FirstName
	user.LastName = updateReq.LastName
	user.UserRole = updateReq.UserRole
	user.ID = updateReq.UserID
	dbErr := model.UserUpdate(user)

	if dbErr != nil {
		log.Println(dbErr)
		RecordLoginAttempt(sess)
		sess.Save(r, w)
		updateResp.StatusCode = constants.StatusCode_500
		updateResp.Message = constants.Msg_500
	} else {
		log.Println("Updated successfully")
		updateResp.StatusCode = constants.StatusCode_200
		updateResp.Message = constants.Msg_200
	}

	ReturnJsonResp(w, updateResp)
}

func makeUserLoginResp(statusCode uint16, msg string, url string) webpojo.UserLoginResp {
	userLoginResp := webpojo.UserLoginResp{}
	userLoginResp.StatusCode = statusCode
	userLoginResp.Message = msg
	userLoginResp.URL = url
	userLoginResp.UserRole = 255
	return userLoginResp
}

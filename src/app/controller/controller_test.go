package controller

import (
	hr "app/route/middleware/httprouterwrapper"
	"app/webpojo"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/julienschmidt/httprouter"

	"app/constants"
	"app/provider"
	"app/shared/database"
	fas "app/shared/files_id_storage"
	"app/shared/jsonconfig"
	"app/shared/session"

	"app/shared/config"
)

var (
	sessionID string
)

// TestController run controller handlers tests
// Tests work with reuqest recorder, so, we handle request by testing handler, and read response via recorder.
func TestController(t *testing.T) {
	time.Sleep(2 * time.Second) // little delay for start tests, because it conflict with provider tests

	// Load config what controller need
	jsonconfig.Load(constants.ConfigFilePath, config.Config)
	database.Connect(config.Database())
	session.Configure(config.Session())

	t.Run("TestBestRates", func(t *testing.T) {
		criterias := [2]string{"best", ""} // test for all items and best items

		for _, criteria := range criterias {
			req, err := http.NewRequest("POST", "/api/admin/rate/list", bytes.NewBuffer([]byte(`{"criteria":"`+criteria+`"}`)))
			if err != nil {
				t.Fatal("fail TestBestRates: ", err)
				return
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(LenderRateList)

			handler.ServeHTTP(rr, req)

			body, err := ioutil.ReadAll(rr.Body)

			log.Println(string(body))
			if err != nil {
				t.Error("fail TestBestRates: " + err.Error())
				return
			}

			if len(body) == 0 {
				t.Error("fail TestBestRates: body is empty")
				return
			}

			if status := getStatusCode(body, rr.Code); status != http.StatusOK {
				t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
				return
			}
		}
	})

	t.Run("TestRegisterCustomer", func(t *testing.T) {

		provider.RemoveCustomerByEmail(constants.TestUserEmail)

		for i := 0; i < 2; i++ { // test twice. For second request expect conflict code
			req, err := http.NewRequest("POST", "/api/customer/register", bytes.NewBuffer([]byte(`
			{"first_name":"John",
			"last_name":"Doe",
			"email":"`+constants.TestUserEmail+`",
			"password":"1qazxsw2",
			"user_role":4
			}`)))
			if err != nil {
				t.Fatal("fail TestRegisterCustomer: ", err)
				return
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(CustomerRegisterPost)

			handler.ServeHTTP(rr, req)

			body, err := ioutil.ReadAll(rr.Body)
			if err != nil {
				t.Error("fail TestRegisterCustomer: " + err.Error())
				return
			}

			log.Println(string(body))

			if i == 0 {
				if status := getStatusCode(body, rr.Code); status != http.StatusOK {
					t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
					return
				}
			}

			if i == 1 {
				if status := getStatusCode(body, rr.Code); status != http.StatusConflict {
					t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusConflict)
					return
				}
			}
		}
	})

	t.Run("LoginCustomer", func(f *testing.T) {
		passwords := [2]string{"1qazxsw2", "wrong"} // test existing email and wrong. For second response expect 404

		for i := 0; i < 2; i++ {
			req, err := http.NewRequest("POST", "/api/customer/login", bytes.NewBuffer([]byte(`
			{"username":"`+constants.TestUserEmail+`",
			"password":"`+passwords[i]+`"
			}`)))
			if err != nil {
				t.Fatal("fail TestLoginCustomer: ", err)
				return
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(CustomerLoginPost)

			handler.ServeHTTP(rr, req)

			body, err := ioutil.ReadAll(rr.Body)
			if err != nil {
				t.Error("fail TestLoginCustomer: " + err.Error())
				return
			}

			if i == 0 {
				if status := getStatusCode(body, rr.Code); status != http.StatusOK {
					t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
					return
				}
			}

			if i == 1 {
				if status := getStatusCode(body, rr.Code); status != http.StatusNotFound {
					t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
					return
				}
			}
		}
	})

	t.Run("LogoutCustomer", func(f *testing.T) {
		req, err := http.NewRequest("POST", "/api/customer/logout", nil)
		if err != nil {
			t.Fatal("fail TestLogoutCustomer: ", err)
			return
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(CustomerLogoutPost)
		setSession(req, rr)

		handler.ServeHTTP(rr, req)

		body, err := ioutil.ReadAll(rr.Body)
		if err != nil {
			t.Error("fail TestLogoutCustomer: " + err.Error())
			return
		}

		if status := getStatusCode(body, rr.Code); status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
			return
		}
	})

	t.Run("UploadCustomerFile", func(f *testing.T) {
		req, err := NewFileUploadRequest("/api/customer/file/upload", uploadFileFormName, "../../../config"+string(os.PathSeparator)+"encryption_test.txt")
		if err != nil {
			log.Println("error while upload customer file: ", err)
			t.Fatal("fail UploadCustomerFile: ", err)
			return
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(CustomerUploadFilePost)
		setSession(req, rr)

		handler.ServeHTTP(rr, req)

		body, err := ioutil.ReadAll(rr.Body)
		if err != nil {
			t.Error("fail UploadCustomerFile: " + err.Error())
			return
		}

		if status := getStatusCode(body, rr.Code); status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
			return
		}
	})

	t.Run("TestUserFileList", func(f *testing.T) {
		req, err := http.NewRequest("GET", "/api/customer/file/list", nil)
		if err != nil {
			t.Fatal("fail TestUserFileList: ", err)
			return
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(CustomerFilesListGet)
		setSession(req, rr)

		handler.ServeHTTP(rr, req)
		body, err := ioutil.ReadAll(rr.Body)
		if err != nil {
			t.Error("fail TestUserFileList: " + err.Error())
			return
		}

		if len(body) == 0 {
			t.Fatal("fail TestUserFileList: reponse body is nil")
			return
		}

		if status := getStatusCode(body, rr.Code); status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
			return
		}
	})

	t.Run("UserGetFile", func(f *testing.T) {
		fileList, err := provider.GetFileList(TestUserID, sessionID)
		if err != nil {
			t.Fatal(err)
		}

		if len(fileList) == 0 {
			t.Error("file list is void")
			return
		}

		req, err := http.NewRequest("POST", "/api/customer/file/get", bytes.NewBuffer([]byte(`{"id": "`+fileList[0].FileID+`"}`)))
		if err != nil {
			t.Fatal("fail UserGetFile: ", err)
		}

		rr := httptest.NewRecorder()

		handler := hr.Handler(http.HandlerFunc(CustomerServeFileGet))

		router := httprouter.New()
		router.POST("/api/customer/file/get", handler)
		setSession(req, rr)

		router.ServeHTTP(rr, req)
		body, err := ioutil.ReadAll(rr.Body)
		if err != nil {
			t.Error("fail UserGetFile: " + err.Error())
		}

		log.Println("UserGetFile: ", string(body))

		if len(body) == 0 {
			t.Fatal("fail UserGetFile: reponse body is nil")
		}

		if status := getStatusCode(body, rr.Code); status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}
	})

	t.Run("TestServeContent", func(f *testing.T) {
		fileList, err := provider.GetFileList(TestUserID, sessionID)
		if err != nil {
			t.Fatal("fail TestServeContent: ", err)
			return
		}

		if len(fileList) == 0 {
			t.Error("fail TestServeContent: file list is void")
			return
		}

		req, err := http.NewRequest("GET", fileList[0].Link, nil)

		// don't delete it please. This header should be set if we using SSE with customer keys
		// encryption model.
		//
		// req.Header.Set("x-amz-server-side-encryption-customer-key", provider.CreateUsersFileKey(TestUserName))
		// req.Header.Set("x-amz-server-side-encryption-customer-key-MD5", provider.GetMD5Hash(provider.CreateUsersFileKey(TestUserName)))
		// req.Header.Set("x-amz-server-side-encryption-customer-algorithm", "AES256")

		client := &http.Client{}
		response, err := client.Do(req)

		if err != nil {
			t.Error("error while try to download file by link: " + err.Error())
			return
		}

		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			t.Error("fail TestServeContent: error while read file body: " + err.Error())
			return
		}

		log.Printf("%s\n", string(contents))

	})

	t.Run("DeleteCustomerFile", func(f *testing.T) {
		fileList, err := provider.GetFileList(TestUserID, sessionID)
		if err != nil {
			t.Fatal("fail DeleteCustomerFile: ", err)
			return
		}

		if len(fileList) == 0 {
			t.Error(errors.New("fail DeleteCustomerFile: file list is empty"))
			return
		}

		req, err := http.NewRequest("DELETE", "/api/customer/file", bytes.NewBuffer([]byte(`{"id":"`+fileList[0].FileID+`"}`)))
		if err != nil {
			t.Fatal("fail DeleteCustomerFile: ", err)
			return
		}

		realID, err := fas.GetRealID(fileList[0].FileID, sessionID)
		log.Println(err)
		if err != nil {
			t.Fatal("fail DeleteCustomerFile: ", err)
			return
		}

		defer func() {
			err = provider.RemoveFileWithAWS(fmt.Sprint(realID) + filepath.Ext(fileList[0].FileName))
			if err != nil {
				t.Fatal("fail DeleteCustomerFile: error while remove file with AWS: ", err)
				return
			}
		}()

		rr := httptest.NewRecorder()
		handler := hr.Handler(http.HandlerFunc(CustomerFileDelete))

		router := httprouter.New()
		router.DELETE("/api/customer/file", handler)
		setSession(req, rr)

		router.ServeHTTP(rr, req)

		body, err := ioutil.ReadAll(rr.Body)
		if err != nil {
			t.Error("fail DeleteCustomerFile: " + err.Error())
		}

		if status := getStatusCode(body, rr.Code); status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}
	})

	t.Run("TestGetUserProfile", func(f *testing.T) {
		req, err := http.NewRequest("POST", "/api/customer/profile", bytes.NewBuffer([]byte(`{"id": `+fmt.Sprint(TestUserID)+`}`)))
		if err != nil {
			t.Fatal("fail TestGetUserProfile: ", err)
			return
		}

		rr := httptest.NewRecorder()
		handler := hr.Handler(http.HandlerFunc(CustomerProfileGetInfo))

		router := httprouter.New()
		router.POST("/api/customer/profile", handler)
		setSession(req, rr)
		router.ServeHTTP(rr, req)

		body, err := ioutil.ReadAll(rr.Body)
		if err != nil {
			t.Error("fail TestGetUserProfile: " + err.Error())
		}

		log.Println(string(body))

		if status := getStatusCode(body, rr.Code); status != http.StatusOK {
			t.Fatalf("fail TestGetUserProfile: handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
	})

	t.Run("TestUserMessagePost", func(f *testing.T) {
		req, err := http.NewRequest("POST", "/api/user/message", bytes.NewBuffer([]byte(
			`{
				"to_user_id":`+fmt.Sprint(TestUserID)+`,
				"from_user_id":`+fmt.Sprint(TestUserID)+`,
				"subject":"some_subject",
				"content":"lalala"
			}`)))

		if err != nil {
			t.Fatal("fail TestUserMessagePost: ", err)
		}

		rr := httptest.NewRecorder()
		handler := hr.Handler(http.HandlerFunc(UserMessagePost))

		router := httprouter.New()
		router.POST("/api/user/message", handler)
		setSession(req, rr)
		router.ServeHTTP(rr, req)

		body, err := ioutil.ReadAll(rr.Body)
		if err != nil {
			t.Error("fail TestUserMessagePost: " + err.Error())
			return
		}

		log.Println(string(body))

		resp := &webpojo.PostNewMessageResp{}
		err = json.Unmarshal(body, resp)
		if err != nil {
			t.Error("error while unmarshal post new message response: ", err)
			return
		}

		if !resp.NewThreadCreated {
			t.Error("error while TestUserMessagePost: new thread should be created")
			return
		}

		if resp.ThreadID == 0 {
			t.Error("error while TestUserMessagePost: thread ID should not be equal to 0")
			return
		}

		if status := getStatusCode(body, rr.Code); status != http.StatusOK {
			t.Errorf("fail TestUserMessagePost: handler returned wrong status code: got %v want %v", status, http.StatusOK)
			return
		}
	})

	t.Run("TestCustomerThreadsList", func(f *testing.T) {
		req, err := http.NewRequest("POST", "/api/user/message/threads", bytes.NewBuffer([]byte(`{"offset":0, "count":10}`)))
		if err != nil {
			t.Fatal("fail TestCustomerMessageList: ", err)
			return
		}

		rr := httptest.NewRecorder()
		handler := hr.Handler(http.HandlerFunc(MessageThreads))

		router := httprouter.New()
		router.POST("/api/user/message/threads", handler)
		setSession(req, rr)
		router.ServeHTTP(rr, req)

		body, err := ioutil.ReadAll(rr.Body)
		if err != nil {
			t.Error("fail TestCustomerMessageList: " + err.Error())
			return
		}

		log.Println(string(body))
		resp := []*webpojo.MessagesThreadsResp{}
		err = json.Unmarshal(body, &resp)
		if err != nil {
			t.Error("error while test TestCustomerThreadsList: can't unmarshal threads response: ", err)
			return
		}

		if len(resp) != 1 {
			t.Error("error while test TestCustomerThreadsList: response list should be 1")
			return
		}

		if resp[0].Title != "some_subject" {
			t.Error("error while test TestCustomerThreadsList: wrong thread title")
			return
		}

		if resp[0].Content != "lalala" {
			t.Error("error while TestCustomerThreadsList: wrong content")
			return
		}

		if fmt.Sprint(resp[0].FromUserID) != TestUserID {
			t.Error("error while TestCustomerThreadsList: fromUserID is not equal")
			return
		}

		if fmt.Sprint(resp[0].ToUserID) != TestUserID {
			t.Error("error while TestCustomerThreadsList: fromUserID is not equal")
			return
		}

		if status := getStatusCode(body, rr.Code); status != http.StatusOK {
			t.Errorf("fail TestCustomerMessageList: handler returned wrong status code: got %v want %v", status, http.StatusOK)
			return
		}
	})

	t.Run("TestCustomerMessageList", func(f *testing.T) {
		threads, err := provider.ThreadsByUserID(TestUserID, 10, 0)
		if err != nil {
			t.Error("error while post new message: can't get threads by user ID: ", err)
			return
		}

		req, err := http.NewRequest("POST", "/api/customer/message/list", bytes.NewBuffer([]byte(`
			{
				"thread_id":`+fmt.Sprint(threads[0].ID)+`,
				"count":10,
				"offseet":0
			}`)))
		if err != nil {
			t.Fatal("fail TestCustomerMessageList: ", err)
			return
		}

		rr := httptest.NewRecorder()
		handler := hr.Handler(http.HandlerFunc(UserMessageList))

		router := httprouter.New()
		router.POST("/api/customer/message/list", handler)
		setSession(req, rr)
		router.ServeHTTP(rr, req)

		body, err := ioutil.ReadAll(rr.Body)
		if err != nil {
			t.Error("fail TestCustomerMessageList: " + err.Error())
			return
		}

		log.Println(string(body))
		resp := []*webpojo.MessageListResp{}
		err = json.Unmarshal(body, &resp)
		if err != nil {
			t.Error("error while TestCustomerMessageList: can't unmarshal response: ", err)
			return
		}

		if len(resp) != 1 {
			t.Error("error while TestCustomerMessageList: can't unmarshal response")
			return
		}

		if resp[0].ThreadID != threads[0].ID {
			t.Error("error while TestCustomerMessageList: threadID is not equal")
			return
		}

		if resp[0].Content != "lalala" {
			t.Error("error while TestCustomerMessageList: content is not equal")
			return
		}

		if status := getStatusCode(body, rr.Code); status != http.StatusOK {
			t.Errorf("fail TestCustomerMessageList: handler returned wrong status code: got %v want %v", status, http.StatusOK)
			return
		}
	})

	t.Run("TestUserMessagePatch", func(f *testing.T) {
		threads, err := provider.ThreadsByUserID(TestUserID, 10, 0)
		if err != nil {
			t.Error("error while TestUserMessagePatch: can't get threads by user ID: ", err)
			return
		}

		iUserID, err := strconv.Atoi(TestUserID)
		if err != nil {
			t.Error("error while TestUserMessagePatch: can't parse userID: ", err)
			return
		}

		messages, err := provider.GetCustomerMessagesList(threads[0].ID, uint32(iUserID), 100, 0)
		if err != nil {
			t.Fatal("fail TestUserMessagePatch: error while get user's messages list: " + err.Error())
			return
		}

		req, err := http.NewRequest("PATCH", "/api/admin/message", bytes.NewBuffer([]byte(`
			{
				"to_user_id": `+TestUserID+`,
				"from_user_id":`+TestUserID+`,
				"message_id":`+fmt.Sprint(messages[0].MessageID)+`,
				"content":"another_test_message",
				"thread_id":`+fmt.Sprint(threads[0].ID)+`
			}
			`)))
		if err != nil {
			t.Fatal("fail TestUserMessagePatch: ", err)
			return
		}

		rr := httptest.NewRecorder()
		handler := hr.Handler(http.HandlerFunc(AdminMessagePatch))

		router := httprouter.New()
		router.PATCH("/api/admin/message", handler)
		setSession(req, rr)
		router.ServeHTTP(rr, req)

		body, err := ioutil.ReadAll(rr.Body)
		if err != nil {
			t.Error("fail TestUserMessagePatch: " + err.Error())
			return
		}

		log.Println("TestUserMessagePatch: ", string(body))

		if status := getStatusCode(body, rr.Code); status != http.StatusOK {
			t.Errorf("fail TestUserMessagePatch: handler returned wrong status code: got %v want %v", status, http.StatusOK)
			return
		}

		messages, err = provider.GetCustomerMessagesList(threads[0].ID, uint32(iUserID), 100, 0)
		if err != nil {
			t.Fatal("fail TestUserMessagePatch: error while get user's messages list: " + err.Error())
			return
		}

		if messages[0].Content != "another_test_message" {
			t.Error("fail TestUserMessagePatch: test content no edited")
			return
		}
	})

	t.Run("TestSetMessageReaded", func(f *testing.T) {
		threads, err := provider.ThreadsByUserID(TestUserID, 10, 0)
		if err != nil {
			t.Error("error while TestSetMessageReaded: can't get threads by user ID: ", err)
			return
		}

		iUserID, err := strconv.Atoi(TestUserID)
		if err != nil {
			t.Error("error while TestSetMessageReaded: can't parse userID: ", err)
			return
		}

		messages, err := provider.GetCustomerMessagesList(threads[0].ID, uint32(iUserID), 100, 0)
		if err != nil {
			t.Fatal("fail TestSetMessageReaded: error while get user's messages list: " + err.Error())
			return
		}

		req, err := http.NewRequest("PATCH", "/api/user/message", bytes.NewBuffer([]byte(`{"thread_id": `+fmt.Sprint(threads[0].ID)+`, "message_id": `+fmt.Sprint(messages[0].MessageID)+`, "readed":true }`)))
		if err != nil {
			t.Fatal("fail TestCustomerMessagePatch: ", err)
			return
		}

		rr := httptest.NewRecorder()
		handler := hr.Handler(http.HandlerFunc(MarkMessageAsReaded))

		router := httprouter.New()
		router.PATCH("/api/user/message", handler)
		setSession(req, rr)
		router.ServeHTTP(rr, req)

		body, err := ioutil.ReadAll(rr.Body)
		if err != nil {
			t.Error("fail TestCustomerMessagePatch: " + err.Error())
			return
		}

		log.Println("TestCustomerMessagePatch: ", string(body))

		if status := getStatusCode(body, rr.Code); status != http.StatusOK {
			t.Errorf("fail TestCustomerMessagePatch: handler returned wrong status code: got %v want %v", status, http.StatusOK)
			return
		}

		messages, err = provider.GetCustomerMessagesList(threads[0].ID, uint32(iUserID), 100, 0)
		if err != nil {
			t.Fatal("fail TestCustomerMessagePatch: error while get user's messages list: " + err.Error())
		}

		if messages[0].Readed != true {
			t.Error("fail TestCustomerMessagePatch: readed no edited")
		}
	})

	t.Run("TestUserMessageDelete", func(f *testing.T) {
		threads, err := provider.ThreadsByUserID(TestUserID, 10, 0)
		if err != nil {
			t.Error("error while TestUserMessageDelete: can't get threads by user ID: ", err)
			return
		}

		iUserID, err := strconv.Atoi(TestUserID)
		if err != nil {
			t.Error("error while TestUserMessageDelete: can't parse userID: ", err)
			return
		}

		messages, err := provider.GetCustomerMessagesList(threads[0].ID, uint32(iUserID), 100, 0)
		if err != nil {
			t.Fatal("fail TestUserMessageDelete: error while get user's messages list: " + err.Error())
			return
		}

		req, err := http.NewRequest("DELETE", "/api/admin/message", bytes.NewBuffer([]byte(`
			{
				"to_user_id": `+TestUserID+`, 
				"from_user_id": `+TestUserID+`,
				"message_id": `+fmt.Sprint(messages[0].MessageID)+`, 
				"thread_id": `+fmt.Sprint(threads[0].ID)+`
			}`)))
		if err != nil {
			t.Fatal("fail TestUserMessageDelete: ", err)
			return
		}

		rr := httptest.NewRecorder()
		handler := hr.Handler(http.HandlerFunc(AdminMessageDelete))

		router := httprouter.New()
		router.DELETE("/api/admin/message", handler)
		setSession(req, rr)
		router.ServeHTTP(rr, req)

		body, err := ioutil.ReadAll(rr.Body)
		if err != nil {
			t.Error("fail TestUserMessageDelete: " + err.Error())
			return
		}

		log.Println("TestUserMessageDelete: ", string(body))

		if status := getStatusCode(body, rr.Code); status != http.StatusOK {
			t.Errorf("fail TestUserMessageDelete: handler returned wrong status code: got %v want %v", status, http.StatusOK)
			return
		}

		messages, err = provider.GetCustomerMessagesList(threads[0].ID, uint32(iUserID), 100, 0)
		if err != nil {
			t.Fatal("error while get user's messages list: " + err.Error())
		}

		if len(messages) != 0 {
			t.Fatal("message not deleted")
		}
	})

	t.Run("TestPatchUserProfile", func(f *testing.T) {
		req, err := http.NewRequest("PATCH", "/api/customer/profile", bytes.NewBuffer([]byte(`{"user_id": "`+TestUserID+`", "username":"wer", "password":"opopop", "email":"ivanov@gmail.com", "mailing_address":"kirovohrad", "phone":"0999858858"}`)))
		if err != nil {
			t.Fatal("fail TestPatchUserProfile: ", err)
			return
		}

		rr := httptest.NewRecorder()
		handler := hr.Handler(http.HandlerFunc(CustomerProfilePatch))

		router := httprouter.New()
		router.PATCH("/api/customer/profile", handler)
		setSession(req, rr)
		router.ServeHTTP(rr, req)

		body, err := ioutil.ReadAll(rr.Body)
		if err != nil {
			t.Error("fail TestPatchUserProfile: " + err.Error())
		}

		log.Println("TestPatchUserProfile: ", string(body))

		if status := getStatusCode(body, rr.Code); status != http.StatusOK {
			t.Errorf("fail TestPatchUserProfile: handler returned wrong status code: got %v want %v", status, http.StatusOK)
			return
		}

		profile, err := provider.GetCustomerProfile(TestUserID)
		if err != nil {
			t.Error("fail TestPatchUserProfile: ", err)
			return
		}

		if profile.Email != "ivanov@gmail.com" {
			t.Error("fail TestPatchUserProfile: user's email not patched")
			return
		}

		if profile.MailingAddress != "kirovohrad" {
			t.Error("fail TestPatchUserProfile: user's mailing address not patched")
			return
		}

		if profile.Phone != "0999858858" {
			t.Error("fail TestPatchUserProfile: user's phone not patched")
			return
		}

		err = provider.RemoveCustomerByEmail("ivanov@gmail.com")
		if err != nil {
			t.Error("fail TestPatchUserProfile: fail TestRemoveCustomerByEmail: " + err.Error())
			return
		}
	})

	t.Run("TestFastQuotePost", func(f *testing.T) {
		req, err := http.NewRequest("POST", "/api/public/fast_quote", bytes.NewBuffer([]byte(`{"loan_amounte":12, "property_value":12, "property_zip":12, "loan_purpose":"hello", "credit_score":"world"}`)))
		if err != nil {
			t.Fatal("fail TestFastQuotePost: ", err)
			return
		}

		rr := httptest.NewRecorder()
		handler := hr.Handler(http.HandlerFunc(FastQuotePost))

		router := httprouter.New()
		router.POST("/api/public/fast_quote", handler)
		setSession(req, rr)
		router.ServeHTTP(rr, req)

		body, err := ioutil.ReadAll(rr.Body)
		if err != nil {
			t.Error("fail TestFastQuotePost: " + err.Error())
			return
		}

		log.Println("fail TestFastQuotePost: ", string(body))

		if status := getStatusCode(body, rr.Code); status != http.StatusOK {
			t.Errorf("fail TestFastQuotePost: handler returned wrong status code: got %v want %v", status, http.StatusOK)
			return
		}

		rate := &[]*webpojo.LenderRatePojo{}
		err = json.Unmarshal(body, rate)
		if err != nil {
			log.Println("fail TestFastQuotePost: can't unmarshal request: ", err)
			t.Error("fail TestFastQuotePost: can't unmarshal request: ", err)
		}
	})
}

/****************** UTILS FUNC**************************/
func NewFileUploadRequest(uri string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Println("error while open file for file upload request: ", err)
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		log.Println("error while crate form file for file upload request: ", err)
		return nil, err
	}
	_, err = io.Copy(part, file)

	err = writer.Close()
	if err != nil {
		log.Println("error while close file at file upload request: ", err)
		return nil, err
	}

	user, err := provider.GetUserByEmail(constants.TestUserEmail)
	if err != nil {
		log.Println("error while get user by email: ", err)
		return nil, err
	}

	setTestUserID(fmt.Sprint(user.ID))
	req, err := http.NewRequest("POST", uri, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}

func setSession(r *http.Request, rr *httptest.ResponseRecorder) {
	sess := session.Instance(r)
	sess.Values[UserID] = TestUserID
	sess.Values[UserName] = TestUserName
	sess.Values[UserRole] = constants.CustomerRole

	sessionID = sess.ID
	sess.Save(r, rr)
}

func getStatusCode(body []byte, code int) int {
	if code != http.StatusOK {
		return code
	}

	strBody := string(body)

	if strings.Contains(strBody, "200") && strings.Contains(strBody, "statusCode") {
		return 200
	}

	if strings.Contains(strBody, "400") && strings.Contains(strBody, "statusCode") {
		return 400
	}

	if strings.Contains(strBody, "401") && strings.Contains(strBody, "statusCode") {
		return 401
	}

	if strings.Contains(strBody, "404") && strings.Contains(strBody, "statusCode") {
		return 404
	}

	if strings.Contains(strBody, "409") && strings.Contains(strBody, "statusCode") {
		return 409
	}

	if strings.Contains(strBody, "429") && strings.Contains(strBody, "statusCode") {
		return 429
	}

	if strings.Contains(strBody, "500") && strings.Contains(strBody, "statusCode") {
		return 500
	}

	return code
}

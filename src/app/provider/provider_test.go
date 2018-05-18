package provider

import (
	"app/shared/jsonconfig"
	"app/shared/passhash"
	"errors"
	"fmt"
	"testing"

	"app/constants"
	"app/model"
	"app/shared/config"
	"app/shared/database"
	"app/webpojo"
)

const (
	customerMessageSubject       = "subject"
	customerMessageContent       = "text"
	customerSecondMessageContent = "second_text"

	customerMessageSubjectChanged = "subject+changed"
	customerMessageContentChanged = "text+changed"
)

func TestProvider(t *testing.T) {

	var testUser *model.User

	// Load the configuration file
	jsonconfig.Load(constants.ConfigFilePath, config.Config)
	database.Connect(config.Database())

	RemoveCustomerByEmail(constants.TestUserEmail)

	t.Run("TestRegisterNewCustomer", func(t *testing.T) {
		regReq := &webpojo.UserCreateReq{UserPojo: webpojo.UserPojo{
			UserID:    0,
			FirstName: "John",
			LastName:  "Doe",
			Email:     constants.TestUserEmail,
			Password:  "1qazxsw2",
			UserRole:  4,
		},
		}

		password, err := passhash.HashString(regReq.Password)
		if err != nil {
			t.Error("fail of TestRegisterNewCustomer: " + err.Error())
		}

		err = RegisterNewCustomer(regReq, password)
		if err != nil {
			t.Error("fail of TestRegisterNewCustomer: " + err.Error())
		}
	})

	t.Run("TestIsUserExist", func(t *testing.T) {
		exist, err := IsUserExist(constants.TestUserEmail)
		if err != nil {
			t.Error("fail TestIsUserExist: " + err.Error())
		}

		if !exist {
			t.Error("fail TestIsUserExist: user shuold be exist")
		}
	})

	t.Run("TestGetUserByEmail", func(t *testing.T) {
		user, err := GetUserByEmail(constants.TestUserEmail)
		if err != nil {
			t.Error("fail TestGetUserByEmail: " + err.Error())
		}

		testUser = user

		if user == nil {
			t.Error("fail TestGetUserByEmail: user shuold be exist")
		}

		if user.FirstName != "John" || user.LastName != "Doe" {
			t.Error("fail TestGetUserByEmail: wrong user name")
		}
	})

	t.Run("TestUpdateCache", func(t *testing.T) {
		err := UpdateCache("update_best_rates")
		if err != nil {
			t.Error("fail TestUpdateCache: " + err.Error())
		}
	})

	t.Run("TestBestLenderRates", func(t *testing.T) {
		list, err := GetBestLenderRatesList()

		if err != nil {
			t.Error("fail TestBestLenderRates: " + err.Error())
		}

		if len(list) != 3 {
			t.Error("fail TestBestLenderRates: wrong list length")
		}

		if list[0].Interest != "3.23" {
			t.Error("fail TestBestLenderRates: wrong interest rate")
			return
		}

		if list[1].Interest != "3.63" {
			t.Error("fail TestBestLenderRates: wrong interest rate")
			return
		}

		if list[2].Interest != "2.95" {
			t.Error("fail TestBestLenderRates: wrong interest rate")
			return
		}
	})

	t.Run("TestSaveFileDataInDB", func(t *testing.T) {
		id, err := SaveFileDataInDB("testFile", "raw", fmt.Sprint(testUser.ID))
		if err != nil {
			t.Error("fail TestSaveFileDataInDB: " + err.Error())
		}

		if id == 0 {
			t.Error("fail TestSaveFileDataInDB: bad id")
		}
	})

	t.Run("TestGetFileList", func(t *testing.T) {
		list, err := GetFileList(fmt.Sprint(testUser.ID), "")
		if err != nil {
			t.Error("fail TestGetFileList: " + err.Error())
		}

		if len(list) == 0 {
			t.Error("fail TestGetFileList: list is empty")
		}
	})

	t.Run("TestPostNewMessage", func(t *testing.T) {
		resp, err := PostNewMessage(fmt.Sprint(testUser.ID), &webpojo.MessagePostReq{
			FromUserID: testUser.ID,
			ToUserID:   testUser.ID,
			Subject:    customerMessageSubject,
			Content:    customerMessageContent,
		})

		if err != nil {
			t.Error("error while TestPostNewMessage: PostNewMessage: " + err.Error())
			return
		}

		threads, err := ThreadsByUserID(fmt.Sprint(testUser.ID), 10, 0)
		if err != nil {
			t.Error("error while post new message: can't get threads by user ID: ", err)
			return
		}

		if len(threads) != 1 {
			t.Error("error while post new message: len of threads is wrong: ", len(threads))
			return
		}

		if threads[0].Title != customerMessageSubject {
			t.Error("error while post new message: wrong subject")
			return
		}

		if threads[0].Content != customerMessageContent {
			t.Error("error while post new message: thread last message is wrong")
			return
		}

		if !resp.NewThreadCreated {
			t.Error("error while post new message: new thread should be created")
			return
		}

		if resp.ThreadID != threads[0].ID {
			t.Error("error while post new message: thread ID in response and in threads list is not equal")
			return
		}

		resp, err = PostNewMessage(fmt.Sprint(testUser.ID), &webpojo.MessagePostReq{
			FromUserID: testUser.ID,
			ToUserID:   testUser.ID,
			ThreadID:   threads[0].ID,
			Content:    customerSecondMessageContent,
		})

		if resp.NewThreadCreated {
			t.Error("errpr while post new message: new thread should not be created")
			return
		}

		messages, err := GetCustomerMessagesList(threads[0].ID, testUser.ID, 10, 0)
		if err != nil {
			t.Error("error while TestPostNewMessage: GetCustomerMessagesList: " + err.Error())
			return
		}

		if len(messages) != 2 {
			t.Error("error while TestPostNewMessage: GetCustomerMessagesList: messages list is empty" + err.Error())
			return
		}

		AccertEqual(t, "TestPostNewMessage - content", customerMessageContent, messages[0].Content)
		AccertEqual(t, "TestPostNewMessage - content", customerSecondMessageContent, messages[1].Content)
	})

	t.Run("TestPatchMessage", func(t *testing.T) {
		threads, err := ThreadsByUserID(fmt.Sprint(testUser.ID), 10, 0)
		if err != nil {
			t.Error("error while TestPatchMessage: can't get threads by user ID: ", err)
			return
		}

		messages, err := GetCustomerMessagesList(threads[0].ID, testUser.ID, 10, 0)
		if err != nil {
			t.Error("error while TestPatchMessage: GetCustomerMessagesList: " + err.Error())
			return
		}

		err = PatchMessage(&webpojo.MessagePatchReq{
			ToUserID:   messages[0].ToUserID,
			FromUserID: messages[0].FromUserID,
			MessageID:  messages[0].MessageID,
			ThreadID:   messages[0].ThreadID,
			Content:    customerMessageContentChanged,
		})

		if err != nil {
			t.Error("error while TestPatchMessage: PatchMessage: " + err.Error())
			return
		}

		messages, err = GetCustomerMessagesList(messages[0].ThreadID, testUser.ID, 10, 0)
		if err != nil {
			t.Error("error while TestPatchMessage: GetCustomerMessagesList: " + err.Error())
			return
		}

		if len(messages) == 0 {
			t.Error("error while TestPatchMessage: GetCustomerMessagesList: messages list is empty" + err.Error())
			return
		}

		AccertEqual(t, "TestPatchMessage - content", customerMessageContentChanged, messages[0].Content)
	})

	t.Run("TestMarkMessageReaded", func(t *testing.T) {
		threads, err := ThreadsByUserID(fmt.Sprint(testUser.ID), 10, 0)
		if err != nil {
			t.Error("error while TestMarkMessageReaded: can't get threads by user ID: ", err)
			return
		}

		messages, err := GetCustomerMessagesList(threads[0].ID, testUser.ID, 10, 0)
		if err != nil {
			t.Error("error while TestMarkMessageReaded: GetCustomerMessagesList: " + err.Error())
			return
		}

		err = MarkMessageReaded(fmt.Sprint(testUser.ID), &webpojo.UserMessagePatchReq{
			MessageID: messages[0].MessageID,
			ThreadID:  messages[0].ThreadID,
			Readed:    true,
		})
		if err != nil {
			t.Error("error while TestMarkMessageReaded: MarkMessageReaded: " + err.Error())
			return
		}

		messages, err = GetCustomerMessagesList(messages[0].ThreadID, testUser.ID, 10, 0)
		if err != nil {
			t.Error("error while TestMarkMessageReaded: GetCustomerMessagesList: " + err.Error())
			return
		}

		if len(messages) == 0 {
			t.Error("error while TestMarkMessageReaded: GetCustomerMessagesList: messages list is empty" + err.Error())
			return
		}

		if !messages[0].Readed {
			t.Error("error while TestMarkMessageReaded: massage shpuld be readed")
			return
		}
	})

	t.Run("TestDeleteMessage", func(t *testing.T) {
		threads, err := ThreadsByUserID(fmt.Sprint(testUser.ID), 10, 0)
		if err != nil {
			t.Error("error while TestDeleteMessage: can't get threads by user ID: ", err)
			return
		}

		messages, err := GetCustomerMessagesList(threads[0].ID, testUser.ID, 10, 0)
		if err != nil {
			t.Error("error while TestDeleteMessage: GetCustomerMessagesList: " + err.Error())
			return
		}

		err = DeleteMessage(testUser.ID, messages[0].MessageID, messages[0].ThreadID)
		if err != nil {
			t.Error("error while TestDeleteMessage: DeleteMessage: " + err.Error())
			return
		}

		messages, err = GetCustomerMessagesList(threads[0].ID, testUser.ID, 10, 0)
		if err != nil {
			t.Error("error while TestDeleteMessage: DeleteMessage: " + err.Error())
			return
		}

		if len(messages) != 1 {
			t.Error("error while TestDeleteMessage: DeleteMessage: messages list is not empty")
			return
		}
	})

	t.Run("TestGetCustomerProfile", func(t *testing.T) {
		result, err := GetCustomerProfile(fmt.Sprint(testUser.ID))
		if err != nil {
			t.Error(err)
			return
		}

		if result.FirstName != "John" || result.LastName != "Doe" {
			t.Error("fail TestGetCustomerProfile: wrong user name")
			return
		}
	})

	t.Run("TestPatchCustomerProfileInfo", func(t *testing.T) {

		password, _ := passhash.HashString("password")
		patchInfo := &webpojo.CustomerInfoPojoPatchReq{
			UserID:         fmt.Sprint(testUser.ID),
			UserName:       "jdoe",
			Password:       password,
			Email:          "johndoecool@gmail.com",
			MailingAddress: "laplandia",
			Phone:          "911",
		}

		err := PatchCustomerProfileInfo(patchInfo)
		if err != nil {
			t.Error(err)
			return
		}

		result, err := GetCustomerProfile(fmt.Sprint(testUser.ID))
		if err != nil {
			t.Error(err)
			return
		}

		if result == nil {
			t.Error("get customer profile result is null")
			return
		}

		AccertEqual(t, "userName", result.UserName, patchInfo.UserName)
		AccertEqual(t, "email", result.Email, patchInfo.Email)
		AccertEqual(t, "mailAddress", result.MailingAddress, patchInfo.MailingAddress)
		AccertEqual(t, "phone", result.Phone, patchInfo.Phone)

		user, err := GetUserByEmail("johndoecool@gmail.com")
		if err != nil {
			t.Error("fail TestGetUserByEmail: " + err.Error())
			return
		}

		if user == nil {
			t.Error("user model is nil")
			return
		}

		AccertEqual(t, "password", user.Password, patchInfo.Password)
	})

	t.Run("TestRemoveCustomerByEmail", func(t *testing.T) {
		err := RemoveCustomerByEmail(constants.TestOtherUserEmail)
		if err != nil {
			t.Error("fail TestRemoveCustomerByEmail: " + err.Error())
			return
		}
	})

	t.Run("TestGetAmazonCredentials", func(t *testing.T) {
		c, err := getAmazonCredentials()
		if err != nil {
			t.Error(err)
			return
		}

		if c == nil {
			t.Error(errors.New("error while get amazon credentials - credentials is nil"))
			return
		}
	})

	t.Run("TestGetAmazonSession", func(t *testing.T) {
		s, err := getAmazonSession()
		if err != nil {
			t.Error(err)
			return
		}

		if s == nil {
			t.Error(errors.New("error while get amazon session - credentials is nil"))
			return
		}
	})

	t.Run("TestGetFastQuote", func(t *testing.T) {
		rate, err := GetFastQuote(&webpojo.FastQuoteReq{})
		if err != nil {
			t.Error(errors.New("fail TestGetFastQuote: " + err.Error()))
			return
		}

		if rate == nil {
			t.Error(errors.New("fail TestGetFastQuote: rate is nil"))
			return
		}
	})
}

func AccertEqual(t *testing.T, name, first, next string) {
	if first != next {
		t.Error(name + " not equal")
	}
}

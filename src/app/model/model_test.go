package model

import (
	"app/constants"
	"app/shared/passhash"
	"fmt"

	"app/shared/config"
	"app/shared/database"
	"app/shared/jsonconfig"
	"app/webpojo"
	"testing"
)

const (
	testUserEmail        = "jane.dou@gmail.com"
	testUserEmailChanged = "jane.dou+changed@gmail.com"

	userPassword        = "password"
	userPasswordChanged = "password+changed"

	customerMessageSubject = "subject"
	customerMessageContent = "text"

	customerMessageSubjectChanged = "subject+changed"
	customerMessageContentChanged = "text+changed"
)

func TestModel(t *testing.T) {

	// Load the configuration file
	jsonconfig.Load(constants.ConfigFilePath, config.Config)
	database.Connect(config.Database())

	UserRemoveByEmail(testUserEmailChanged)

	t.Run("TestUserCreateWithRole", func(t *testing.T) {
		hashedPassword, err := passhash.HashString(userPassword)
		if err != nil {
			t.Error("password hashing failed")
			return
		}

		err = UserCreateWithRole("John", "Doe", testUserEmail, hashedPassword, 4)
		if err != nil {
			t.Error("error while test UserCreateWithRole: " + err.Error())
			return
		}

		user, err := UserByEmail(testUserEmail)
		if err != nil {
			t.Error("error while TestUserCreateWithRole: can't get user by email: " + err.Error())
			return
		}

		AccertEqual(t, "TestUserCreateWithRole - first name", "John", user.FirstName)
		AccertEqual(t, "TestUserCreateWithRole - last name", "Doe", user.LastName)
		AccertEqual(t, "TestUserCreateWithRole - email", testUserEmail, user.Email)
		compairPasswords(t, user.Password, hashedPassword)
	})

	t.Run("TestCheckIsUserExist", func(t *testing.T) {
		user := getUserByEmail(t, testUserEmail)

		exist, err := CheckIsUserExist(user.ID)
		if err != nil {
			t.Error("error while TestCheckIsUserExist: " + err.Error())
		}

		if !exist {
			t.Error("error while TestCheckIsUserExist: expect true, existing user")
		}
	})

	t.Run("TestUserByID", func(t *testing.T) {
		user := getUserByEmail(t, testUserEmail)

		userByID, err := UserByID(fmt.Sprint(user.ID))
		if err != nil {
			t.Error("error while TestUserByID: " + err.Error())
		}

		AccertEqual(t, "TestUserByID - email", user.Email, userByID.Email)
	})

	t.Run("TestUserUpdate", func(t *testing.T) {
		user := getUserByEmail(t, testUserEmail)

		user.Email = testUserEmailChanged
		user.FirstName = "SuperJane"
		user.LastName = "SuperDoe"
		hashedPassword, err := passhash.HashString(userPasswordChanged)
		if err != nil {
			t.Error("error while TestUserUpdate: password hashing failed")
			return
		}

		user.Password = hashedPassword

		err = UserUpdate(*user)
		if err != nil {
			t.Error("error while TestUserUpdate: " + err.Error())
			return
		}

		*user, err = UserByEmail(testUserEmailChanged)
		if err != nil {
			t.Error("error while test TestUserUpdate: can't get changed user by email: " + err.Error())
			return
		}

		AccertEqual(t, "TestUserUpdate - email", testUserEmailChanged, user.Email)
		AccertEqual(t, "TestUserUpdate - first name", "SuperJane", user.FirstName)
		AccertEqual(t, "TestUserUpdate - last name", "SuperDoe", user.LastName)
		compairPasswords(t, user.Password, hashedPassword)
	})

	t.Run("TestCustomerInsert", func(t *testing.T) {
		user := getUserByEmail(t, testUserEmailChanged)

		customerInfo := &Customer{}

		customerInfo.UserID.Scan(user.ID)
		customerInfo.UserName.Scan("superJane2000")
		customerInfo.Password = userPassword
		customerInfo.Email = testUserEmail
		customerInfo.MailingAddress.Scan("Jane's home")
		customerInfo.Phone.Scan("+380999658641")

		err := CustomerInsert(customerInfo)
		if err != nil {
			t.Error("error while TestCustomerInsert: " + err.Error())
			return
		}

		customer, err := RawCustomerByUserID(fmt.Sprint(user.ID))
		if err != nil {
			t.Error("error while TestCustomerInsert: error while get customer info by User ID: " + err.Error())
			return
		}

		*user, err = UserByID(fmt.Sprint(user.ID))
		if err != nil {
			t.Error("error while TestCustomerInsert: error while get user by ID: " + err.Error())
			return
		}

		if customer == nil {
			t.Error("error while TestCustomerInsert: customer struct is nil")
			return
		}

		AccertEqual(t, "TestCustomerInsert - mailing address", customer.MailingAddress, customerInfo.MailingAddress.String)
		AccertEqual(t, "TestCustomerInsert - phone", customer.Phone, customerInfo.Phone.String)
		AccertEqual(t, "TestCustomerInsert - userName", customer.UserName, customerInfo.UserName.String)
		compairPasswords(t, user.Password, customerInfo.Password)
	})

	t.Run("TestRawCustomerByUserID", func(t *testing.T) {
		user := getUserByEmail(t, testUserEmail)

		customer, err := RawCustomerByUserID(fmt.Sprint(user.ID))
		if err != nil {
			t.Error("error while TestRawCustomerByUserID: " + err.Error())
			return
		}

		if customer == nil {
			t.Error("error while TestRawCustomerByUserID: customer struct is nil")
			return
		}
	})

	t.Run("TestCustomerByEmail", func(t *testing.T) {
		customer, err := CustomerByEmail(testUserEmail)
		if err != nil {
			t.Error("error while TestCustomerByEmail: " + err.Error())
			return
		}

		if customer == nil {
			t.Error("error while TestCustomerByEmail: customer struct is nil")
			return
		}
	})

	t.Run("TestCustomerUpdate", func(t *testing.T) {
		user := getUserByEmail(t, testUserEmail)

		hashedPassword, err := passhash.HashString(userPasswordChanged)
		if err != nil {
			t.Error("password hashing failed")
			return
		}

		customerInfo := &Customer{}

		customerInfo.UserID.Scan(user.ID)
		customerInfo.UserName.Scan("superJane2018")
		customerInfo.Password = hashedPassword
		customerInfo.Email = testUserEmailChanged
		customerInfo.MailingAddress.Scan("Jane's another home")
		customerInfo.Phone.Scan("+380999999999")
		customerInfo.FirstName = "John"
		customerInfo.LastName = "Doe"

		err = CustomerUpdate(customerInfo)
		if err != nil {
			t.Error("error whileupdate customer profile info")
		}

		customer, err := RawCustomerByUserID(fmt.Sprint(user.ID))
		if err != nil {
			t.Error("error while TestCustomerUpdate: " + err.Error())
			return
		}

		if customer == nil {
			t.Error("error while TestCustomerUpdate: customer struct is nil")
			return
		}

		user = getUserByEmail(t, testUserEmailChanged)
		AccertEqual(t, "TestCustomerUpdate - user name", customer.UserName, customerInfo.UserName.String)
		compairPasswords(t, user.Password, hashedPassword)
		AccertEqual(t, "TestCustomerUpdate - mailing address", customer.MailingAddress, customerInfo.MailingAddress.String)
		AccertEqual(t, "TestCustomerUpdate - phone", customer.Phone, customerInfo.Phone.String)
		AccertEqual(t, "TestCustomerUpdate - userName", customer.UserName, customerInfo.UserName.String)
	})

	t.Run("TestThreadCreate", func(t *testing.T) {
		userID := getUserByEmail(t, testUserEmailChanged).ID

		threadID, err := ThreadCreate("some_title", userID, userID)
		if err != nil {
			t.Error("error while create new thread: ", err)
			return
		}

		if threadID == 0 {
			t.Error("error while create new thread: thread ID is nil")
			return
		}

		t.Run("TestMessageCreate", func(t *testing.T) {
			userID := getUserByEmail(t, testUserEmailChanged).ID

			err := MessageCreate(userID, userID, threadID, customerMessageContent)
			if err != nil {
				t.Error("error while TestMessageCreate: error while message create: " + err.Error())
				return
			}

			messages, err := MessagesByUserID(threadID, userID, 10, 0)
			if err != nil {
				if err != nil {
					t.Error("error while TestMessageCreate: error while get user's messages by ID: " + err.Error())
					return
				}
			}

			if len(messages) != 1 {
				t.Error("error while TestMessageCreate: len(messages) is not expected")
				return
			}

			AccertEqual(t, "TestMessageCreate - content", messages[0].Content, customerMessageContent)
		})

		threads, err := ThreadsByUserID(fmt.Sprint(userID), 10, 0)
		if err != nil {
			t.Error("error while create new thread: can't get thread list: ", err)
			return
		}

		if len(threads) != 1 {
			t.Error("error while create thread: len of thread list is wrong: ", len(threads))
		}

		if threads[0].Content != customerMessageContent {
			t.Error("error while get threads info: content of last message is wrong")
			return
		}
	})

	t.Run("TestMessageUpdate", func(t *testing.T) {
		userID := getUserByEmail(t, testUserEmailChanged).ID

		threads, err := ThreadsByUserID(fmt.Sprint(userID), 10, 0)
		if err != nil {
			t.Error("error while TestMessageUpdate: error while get threads: " + err.Error())
		}

		messages, err := MessagesByUserID(threads[0].ID, userID, 10, 0)
		if err != nil {
			t.Error("error while TestMessageUpdate: error while get user's messages by ID: " + err.Error())
			return
		}

		err = MessageUpdate(customerMessageContentChanged, userID, messages[0].ID, threads[0].ID)
		if err != nil {
			t.Error("error while TestMessageUpdate: error while message create: " + err.Error())
			return
		}

		messages, err = MessagesByUserID(threads[0].ID, userID, 10, 0)
		if err != nil {
			if err != nil {
				t.Error("error while TestMessageUpdate: error while get user's messages by ID: " + err.Error())
				return
			}
		}

		AccertEqual(t, "TestMessageUpdate - content", messages[0].Content, customerMessageContentChanged)
	})

	t.Run("TestMarkMessageReadedUpdate", func(t *testing.T) {
		userID := getUserByEmail(t, testUserEmailChanged).ID

		threads, err := ThreadsByUserID(fmt.Sprint(userID), 10, 0)
		if err != nil {
			t.Error("error while set message readed: error while get threads: ", err)
			return
		}

		messages, err := MessagesByUserID(threads[0].ID, userID, 10, 0)
		if err != nil {
			if err != nil {
				t.Error("error while TestMarkMessageReadedUpdate: error while get user's messages by ID: " + err.Error())
				return
			}
		}

		err = MarkMessageReadedUpdate(true, userID, messages[0].ID, threads[0].ID)
		if err != nil {
			t.Error("error while TestMarkMessageReadedUpdate: error while message create: " + err.Error())
			return
		}

		messages, err = MessagesByUserID(threads[0].ID, userID, 10, 0)
		if err != nil {
			if err != nil {
				t.Error("error while TestMarkMessageReadedUpdate: error while get user's messages by ID: " + err.Error())
				return
			}
		}

		if !messages[0].Readed {
			t.Error("error while TestMarkMessageReadedUpdate: set messages readed: message is not readed")
		}
	})

	t.Run("TestMessageDelete", func(t *testing.T) {
		userID := getUserByEmail(t, testUserEmailChanged).ID
		if userID == 0 {
			t.Error("error while TestMessageDelete: userID is empty")
		}

		threads, err := ThreadsByUserID(fmt.Sprint(userID), 10, 0)
		if err != nil {
			t.Error("error while TestMessageDelete: error while get threads: " + err.Error())
		}

		messages, err := MessagesByUserID(threads[0].ID, userID, 10, 0)
		if err != nil {
			if err != nil {
				t.Error("error while TestMessageDelete: error while get user's messages by ID: " + err.Error())
				return
			}
		}

		if len(messages) != 1 {
			t.Error("error while TestMessageDelete: len(messages) unexpected")
			return
		}

		err = MessageDelete(userID, messages[0].ID, threads[0].ID)
		if err != nil {
			t.Error("error while test TestMessageDelete: error while message delete: " + err.Error())
			return
		}

		messages, err = MessagesByUserID(threads[0].ID, userID, 10, 0)
		if err != nil {
			if err != nil {
				t.Error("error while test TestMessageDelete: error while get user's messages by ID: " + err.Error())
				return
			}
		}

		if len(messages) != 0 {
			t.Error("error while test TestMessageDelete: error while get user's messages by ID: " + err.Error())
			return
		}
	})

	t.Run("TestFastQuoteRate", func(t *testing.T) {
		rate, err := FastQuoteRate(&webpojo.FastQuoteReq{})
		if err != nil {
			t.Error("error while test TestFastQuoteRate: error while get fast quote rate: " + err.Error())
			return
		}

		if rate == nil {
			t.Error("error while test TestFastQuoteRate: rate is nil")
			return
		}
	})

}

//*******************************************************************************************
//Utils funcs
//*******************************************************************************************

func getUserByEmail(t *testing.T, email string) *User {
	user, err := UserByEmail(email)
	if err != nil {
		t.Fatal("error while get user by email: can't get user by email: " + email + " - " + err.Error())
		return nil
	}

	return &user
}

func compairPasswords(t *testing.T, currentPasshash, password string) bool {
	hashedPassword, err := passhash.HashString(password)
	if err != nil {
		t.Error("password hashing failed")
		return false
	}

	return hashedPassword == currentPasshash
}

func AccertEqual(t *testing.T, name, first, next string) {
	if first != next {
		t.Error(name + " not equal: " + first + " - " + next)
	}
}

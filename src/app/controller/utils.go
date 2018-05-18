package controller

import (
	"errors"
	"strconv"

	"github.com/gorilla/sessions"
)

var (
	// TestUserID need for testing file uploading when coockies is not set
	TestUserID string
	// TestUserName need for coockies in controller tests
	TestUserName = "testUser"
)

func getUserID(sess *sessions.Session) string {
	return sess.Values[UserID].(string)
}

func getUserName(sess *sessions.Session) string {
	return sess.Values[UserName].(string)
}

func setTestUserID(userID string) {
	TestUserID = userID
}

func getCountOffsetParams(count, offset string) (int, int, error) {
	if count == "" {
		return 0, 0, errors.New("error while get count offset params: count is void")
	}

	if offset == "" {
		return 0, 0, errors.New("error while get count offset params: offset is void")
	}

	intCount, err := strconv.Atoi(count)
	if err != nil {
		return 0, 0, errors.New("error while get count params: " + err.Error())
	}

	intOffset, err := strconv.Atoi(offset)
	if err != nil {
		return 0, 0, errors.New("error while get offset params: " + err.Error())
	}

	return intCount, intOffset, nil
}

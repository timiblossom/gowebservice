package provider

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"log"
)

// GetMD5Hash Getting md5 hash from key string and ecode it with base64
func GetMD5Hash(key string) string {
	h := md5.New()
	h.Write([]byte(key))
	hashedKey := base64.StdEncoding.EncodeToString(h.Sum(nil))
	log.Println("generate new md5 hash of users file key: ", string(hashedKey))

	return hashedKey
}

// CreateUsersFileKey returned 32 bytes key from hashed userName
func CreateUsersFileKey(userName string) string {
	key := []byte(userName)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(userName))

	usersFileName := base64.StdEncoding.EncodeToString(h.Sum(nil))
	if len(usersFileName) > 32 {
		usersFileName = usersFileName[:32]
	} else {
		usersFileName = (usersFileName + usersFileName)[:32]
	}

	log.Println("generate new usersFileKey: ", usersFileName)
	return usersFileName
}

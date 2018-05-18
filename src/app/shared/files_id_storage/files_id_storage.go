package files_id_storage

import (
	"app/constants"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
)

// GetNewFileID storing real file ID in cache, and return fail ID for frontend
func GetNewFileID(realID int, sessID string) string {
	log.Println("get new file ID with real ID=", realID, " and sess ID=", sessID, " len of info=", len([]byte(sessID+fmt.Sprint(realID))))
	failID, err := Encrypt([]byte(sessID+fmt.Sprint(realID)), constants.AESEncryptKey)
	if err != nil {
		log.Println("error while encrypt new fail ID key: " + err.Error())
		panic(err)
	}

	res := base64.URLEncoding.EncodeToString(failID)
	log.Println("new fake file id generates successfully: " + res)
	return res
}

// GetRealID return real ID by fail ID
func GetRealID(failID string, sessID string) (string, error) {
	log.Println("try to get real id from fail: "+failID, " with sessID: ", sessID)
	decoded, err := base64.URLEncoding.DecodeString(failID)
	if err != nil {
		log.Println("error while decode base64 basic file ID data: " + err.Error())
		return "", err
	}

	log.Println("bytes fake ID: ", []byte(decoded))
	realID, err := Decrypt([]byte(decoded), constants.AESEncryptKey)
	if err != nil {
		log.Println("error while decrypt new file ID key: " + err.Error())
		return "", err
	}

	res := strings.TrimPrefix(string(realID), sessID)
	log.Println("decoded real id is: " + res)
	return strings.TrimPrefix(string(realID), sessID), nil
}

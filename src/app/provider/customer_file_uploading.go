package provider

import (
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"app/constants"
	"app/model"
	fas "app/shared/files_id_storage"
	"app/webpojo"
)

// UploadFile create new file and copy data from multipart reuqest
// This func is testing in controller test, because it need multipart form data
func UploadFile(file *multipart.File, fileHeader *multipart.FileHeader, fileID, userName string) error {
	if file == nil {
		return errors.New("error while upload file: multipart.File is nil")
	}

	if fileHeader == nil {
		return errors.New("error while upload file: multipart.FileHeader is nil")
	}

	if fileID == "" || fileID == "0" {
		return errors.New("error while upload file: file id is empty or zero")
	}

	response, err := uploadFileWithAWS(file, fileID, userName)
	if err != nil {
		return errors.New("error while upload files to AWSS3: " + err.Error())
	}

	log.Println("AWSS3 upload response: " + response)
	return nil
}

// GetDecryptedFileFata return decrypted file content
// Using in intrnal files which storing in /upload directory of the project
func GetDecryptedFileFata(fileID, userID string) ([]byte, error) {
	file, err := model.FileByID(userID, fileID)
	if err != nil {
		return nil, errors.New("error while get ecnrypted file data: " + err.Error())
	}

	data, err := fas.ReadFromFile("../../uploaded/" + file.FileName)
	if err != nil {
		return nil, errors.New("error while read ecnrypted file: " + err.Error())
	}

	data, err = fas.Decrypt(data, constants.FilesCryptoKey)
	if err != nil {
		return nil, errors.New("error while decrypt file data: " + err.Error())
	}

	return data, nil
}

// SaveFileDataInDB create new DB row with file id, name etc
func SaveFileDataInDB(fileName, fileType, userID string) (int, error) {
	if fileName == "" {
		return 0, errors.New("error while save file data in DB: file name is empty")
	}

	if fileType == "" {
		return 0, errors.New("error while save file data in DB: file type is empty")
	}

	if userID == "" || userID == "0" {
		return 0, errors.New("error while save file data in DB: user id is empty or zero")
	}

	lastID, err := model.FileCreate(fileName, fileType, userID)
	if err != nil {
		log.Println("error while add new file info to DB: " + err.Error())
		return 0, errors.New("error while add new file info to DB: " + err.Error())
	}

	return lastID, nil
}

// GetFileList return bytes response with list of user files
func GetFileList(userID string, sessID string) ([]*webpojo.UserFile, error) {
	if userID == "" || userID == "0" {
		return nil, errors.New("error while get user's file list: user id empty or zero")
	}

	rawFileList, err := model.FilesByUserID(userID)
	if err != nil {
		return nil, err
	}

	var resultFilesList []*webpojo.UserFile

	for _, v := range rawFileList {
		link, err := getPresignLinkFromAWS(fmt.Sprint(v.ID) + filepath.Ext(v.FileName))
		if err != nil {
			log.Println("error while get presign link from AWSS3: " + err.Error())
			return nil, err
		}

		// in some reason, app escaping  & and / symbols, so, should replase to righ state
		link = strings.Replace(link, "\u0026", "&", -1)
		link = strings.Replace(link, "%2F", "/", -1)
		link = strings.Replace(link, "%3B", ";", -1)

		resultFilesList = append(resultFilesList, &webpojo.UserFile{Link: link, FileName: v.FileName, FileID: fmt.Sprint(fas.GetNewFileID(int(v.ID), sessID))})
	}

	return resultFilesList, nil
}

// GetFile return user file struct
func GetFile(userID string, fileID string, sessID string) (*webpojo.UserFile, error) {
	if fileID == "" {
		return nil, errors.New("error while get file info from DB: fileID is zero")
	}

	if userID == "" || userID == "0" {
		return nil, errors.New("error while save file data in DB: user id is empty or zero")
	}

	fileInfo, err := model.FileByID(userID, fileID)
	if err != nil {
		return nil, err
	}

	var userFile *webpojo.UserFile

	link, err := getPresignLinkFromAWS(fmt.Sprint(fileInfo.ID) + filepath.Ext(fileInfo.FileName))
	if err != nil {
		log.Println("error while get presign link from AWSS3: " + err.Error())
		return nil, err
	}

	// in some reason, app escaping  & and / symbols, so, should replase to righ state
	link = strings.Replace(link, "\u0026", "&", -1)
	link = strings.Replace(link, "%2F", "/", -1)
	link = strings.Replace(link, "%3B", ";", -1)

	userFile = &webpojo.UserFile{Link: link, FileName: fileInfo.FileName, FileID: fas.GetNewFileID(int(fileInfo.ID), sessID)}

	return userFile, nil
}

// FileDelete delete user file from database (WARNING: but not from storage. Storage contain all users files
// after delete for backup)
func FileDelete(userID string, fileID string) error {
	if fileID == "" {
		return errors.New("error while delete user's file: fileID is zero")
	}

	if userID == "" || userID == "0" {
		return errors.New("error while delete user's file: user id is empty or zero")
	}

	err := model.FileDelete(userID, fileID)
	if err != nil {
		log.Println("error while delete user file from DB: " + err.Error())
		return err
	}

	return nil
}

// PhisicalFileDelete remove file from file system
// Not using in api, just like a tool function
func PhisicalFileDelete(filePath string) error {
	err := os.Remove(filePath)
	if err != nil {
		return errors.New("error while phisical remove file: " + err.Error())
	}

	return nil
}

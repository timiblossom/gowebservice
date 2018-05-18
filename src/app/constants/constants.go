package constants

import "os"

// *****************************************************************************
// Global constants
// *****************************************************************************

const (
	StatusCode_200 uint16 = 200
	Msg_200        string = "Success"

	StatusCode_204 uint16 = 204
	Msg_204        string = "No content"

	StatusCode_400 uint16 = 400
	Msg_400        string = "Bad request"

	StatusCode_401 uint16 = 401
	Msg_401        string = "Unauthorized"

	StatusCode_404 uint16 = 404
	Msg_404        string = "Not found"

	StatusCode_408 uint16 = 408
	Msg_408        string = "Request timeout"

	StatusCode_409 uint16 = 409
	Msg_409        string = "Conflict"

	StatusCode_429 uint16 = 429
	Msg_429        string = "Too many requests"

	StatusCode_500 uint16 = 500
	Msg_500        string = "Internal server error"

	StatusCode_505 uint16 = 505
	Msg_505        string = "Version not supported"

	// *****************************************************************************
	// User Role
	// *****************************************************************************
	StaffRole      = 1
	SupervisorRole = 2
	AdminRole      = 3
	CustomerRole   = 4
	DefaultRole    = 0

	// AESEncryptKey
	AESEncryptKey = "hiuh8ab324ddwp"
	// SSEKMSKeyID need for getting encryption/decryption key
	SSEKMSKeyID = "arn:aws:kms:us-east-1:832507808273:key/a9cfda3b1-583c-43sa-3982-9bbad16ee28c4b"
	// ServerSideEncryptionType specified encryption type
	ServerSideEncryptionType = "aws:kms"
	// FilesCryptoKey using for aes files encryption
	FilesCryptoKey = "FRG563KJH$61GhGTAFDADSAYAFAD16632423431"
	// ConfigFilePath get a way to simple config file getting
	ConfigFilePath = "../../../config" + string(os.PathSeparator) + "config.json"
	// TestUserEmail = random email just for testing
	TestUserEmail = "johndoe@gmail.com"
	// TestOtherUserEmail = random email just for testing
	TestOtherUserEmail = "johndoecool@gmail.com"
	// FilesIDLifePeriodMin is expired period for files ids what is send to frontend
	FilesIDLifePeriodMin = 5
	// AWSS3PresigmLinkLifePeriodMin is expired time for links
	AWSS3PresigmLinkLifePeriodMin = 5
)

package provider

import (
	"app/model"
	"app/webpojo"
	"errors"
	"log"
	"strconv"
)

// GetCustomerProfile return response for GET /api/customer/profile/:user_id
func GetCustomerProfile(userID string) (*webpojo.CustomerInfoPojoResp, error) {
	customerData, err := getCustomerProfileInfo(userID)
	if err != nil {
		log.Println("get customer profile info error: " + err.Error())
		return nil, err
	}

	bytedResult, err := getCustomerProfileInfoResponse(customerData)
	if err != nil {
		log.Println("get customer profile info response error: " + err.Error())
		return nil, err
	}

	return bytedResult, err
}

// getCustomerProfileInfo return full customer info
func getCustomerProfileInfo(userID string) (*model.Customer, error) {
	if userID == "" {
		return nil, errors.New("error while get customer profile info: user id is empty")
	}

	customer, err := model.CustomerByUserID(userID)
	if err != nil {
		log.Println("error while get customer by userID: " + err.Error())
		return nil, err
	}

	return customer, nil
}

// getCustomerProfileInfoResponse make webpojo model and marshal db GetCustomerProfileInfo result
func getCustomerProfileInfoResponse(customer *model.Customer) (*webpojo.CustomerInfoPojoResp, error) {
	if customer == nil {
		return nil, errors.New("error while make customer profile info response: customer info is nil")
	}

	return &webpojo.CustomerInfoPojoResp{
		UserID:         uint32(customer.UserID.Int64),
		FirstName:      customer.FirstName,
		LastName:       customer.LastName,
		Email:          customer.Email,
		UserName:       customer.UserName.String,
		MailingAddress: customer.MailingAddress.String,
		Phone:          customer.Phone.String,
	}, nil
}

// PatchCustomerProfileInfo changing customer info accrording to input info
func PatchCustomerProfileInfo(patchInfo *webpojo.CustomerInfoPojoPatchReq) error {
	_, err := model.RawCustomerByUserID(patchInfo.UserID)
	if err != nil && err != model.ErrNoResult {
		log.Println("error while patch customer: " + err.Error())
		return err
	}

	err = insertCustomerProfileInfo(patchInfo, patchInfo.UserID, err != model.ErrNoResult)
	return err
}

func insertCustomerProfileInfo(patchInfo *webpojo.CustomerInfoPojoPatchReq, userID string, isExist bool) error {
	customerInfo, err := getCustomerProfileInfo(userID)
	if err != nil {
		log.Println("error while insert customer patch info: " + err.Error())
		return err
	}

	customerInfo.UserID.Scan(userID)
	customerInfo.UserName.Scan(patchInfo.UserName)
	customerInfo.Password = patchInfo.Password
	customerInfo.Email = patchInfo.Email
	customerInfo.MailingAddress.Scan(patchInfo.MailingAddress)
	customerInfo.Phone.Scan(patchInfo.Phone)

	if isExist {
		return model.CustomerUpdate(customerInfo)
	}

	return model.CustomerInsert(customerInfo)
}

// CheckIsUserExist checkong user exist state
func CheckIsUserExist(userID uint32) (bool, error) {
	return model.CheckIsUserExist(userID)
}

// CheckIsUserExistString checkong user exist state with string arg
func CheckIsUserExistString(userID string) (bool, error) {
	u, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		log.Println("error while check is user exist: can't parse user ID: " + userID)
		return false, err
	}

	return model.CheckIsUserExist(uint32(u))
}

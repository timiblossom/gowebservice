package webpojo

// CustomerInfoPojoResp represents customer info which should be send to frontend
type CustomerInfoPojoResp struct {
	UserID         uint32 `json:"user_id,omitempty"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Email          string `json:"email"`
	UserName       string `json:"user_name"`
	MailingAddress string `json:"mailing_address"`
	Phone          string `json:"phone"`
}

//CustomerInfoPojoPatchReq represents patch customer info
type CustomerInfoPojoPatchReq struct {
	UserID         string `json:"user_id"`
	UserName       string `json:"username"`
	Password       string `json:"password"`
	Email          string `json:"email"`
	MailingAddress string `json:"mailing_address"`
	Phone          string `json:"phone"`
}

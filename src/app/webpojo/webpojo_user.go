package webpojo

type UserPojo struct {
	UserID    uint32 `json:"user_id,omitempty"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password,omitempty"`
	UserRole  uint8  `json:"user_role,omitempty"`
}

const (
	UserStaff      uint16 = 1
	UserSupervisor uint16 = 2
	UserAdmin      uint16 = 3
	UserCustomer   uint16 = 4
)

//UserLoginReq for admin login request
type UserLoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

//UserLoginResp for admin login response
type UserLoginResp struct {
	StatusCode uint16 `json:"statusCode"`
	Message    string `json:"message"`
	URL        string `json:"url"`
	UserPojo
}

//UserCreateReq for registering an admin user
type UserCreateReq struct {
	UserPojo
}

// IDRequest need for sending some ID in request body
type IDRequest struct {
	ID int `json:"id"`
}

// IDRequestString need for sending some ID in request body
type IDRequestString struct {
	ID string `json:"id"`
}

//UserCreateResp for admin user registration's response
type UserCreateResp struct {
	StatusCode uint16 `json:"statusCode"`
	Message    string `json:"message"`
}

type UserListReq struct {
	Criteria string `json:"criteria"`
}

type UserListResp struct {
	StatusCode uint16     `json:"statusCode"`
	Message    string     `json:"message"`
	Users      []UserPojo `json:"users,omitempty"`
}

//UserUpdateReq for registering an admin user
type UserUpdateReq struct {
	UserPojo
}

type UserUpdateResp struct {
	StatusCode uint16 `json:"statusCode"`
	Message    string `json:"message"`
}

type UserDisableReq struct {
	UserID uint32 `json:"user_id"`
}

type UserDisableResp struct {
	StatusCode uint16 `json:"statusCode"`
	Message    string `json:"message"`
}

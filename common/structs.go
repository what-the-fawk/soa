package common

type NewUserInfo struct {
	Login       string `json:"login"`
	Password    string `json:"password"`
	FirstName   string `json:"first_name"`
	SecondName  string `json:"second_name"`
	DateOfBirth string `json:"date_of_birth"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
}

type UserInfo struct {
	FirstName   string `json:"first_name"`
	SecondName  string `json:"second_name"`
	DateOfBirth string `json:"date_of_birth"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
}

type AuthInfo struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

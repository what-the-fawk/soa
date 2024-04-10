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

type PostInfo struct {
	AuthorId         uint64 `json:"authorId"`
	DateOfCreation   string `json:"dateOfCreation"`
	Content          string `json:"content"`
	CommentSectionId uint64 `json:"commentSectionId"`
}

type PostId struct {
	Id uint64 `json:"postId"`
}

type Post struct {
	PostId           uint64 `json:"PostIdId"`
	AuthorId         uint64 `json:"authorId"`
	DateOfCreation   string `json:"dateOfCreation"`
	Content          string `json:"content"`
	CommentSectionId uint64 `json:"commentSectionId"`
}

type PaginationInfo struct {
	PageNumber uint64 `json:"pageNumber"`
	BatchSize  uint32 `json:"batchSize"`
}

package dto

//need to set up for register and login and jwt

type UserResponse struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	ProfileUrl string `json:"profileUrl"`
	Email      string `json:"email"`
	CreatedAt  string `json:"createdAt"`
}

package apptypes

type User struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginReq struct {
	Email	string `json:"email"`
	Password string `json:"password"`
}
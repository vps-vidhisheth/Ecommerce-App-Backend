package credentials

type Credentials struct {
	Email    string ` json:"email"`
	Password string ` json:"password"`
	Role     string `json:"role"`
	Captcha  string `json:"captcha"`
}

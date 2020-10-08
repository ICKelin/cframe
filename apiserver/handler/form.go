package handler

// edge
type AddEdgeForm struct {
	CSPType    int    `json:"csptype"`
	Name       string `json:"name"`
	PublicIP   string `json:"hostaddr"`
	Cidr       string `json:"cidr"`
	ListenAddr string `json:"listenAddr"`
	Comment    string `json:"comment"`
}

type DelEdgeForm struct {
	Name string `json:"name"`
}

// user
type SignupForm struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	About    string `json:"about"`
}

type SigninForm struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// csp
type AddCSPForm struct {
	CSPType      int    `json:"csptype"`
	AccessKey    string `json:"accessKey"`
	AccessSecret string `json:"accessSecret"`
}

type DelCSPForm struct {
	CSPType int `json:"csptype"`
}

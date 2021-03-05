package handler

// edge
type AddEdgeForm struct {
	CSPType    int    `json:"csptype"`
	Name       string `json:"name"`
	PublicIP   string `json:"hostaddr"`
	Cidr       string `json:"cidr"`
	PublicPort int32  `json:"publicPort"`
	Comment    string `json:"comment"`
}

type DelEdgeForm struct {
	Name string `json:"name"`
}

type GetStatForm struct {
	Name      string `json:"name"`
	From      int64  `json:"from"`
	Count     int32  `json:"to"`
	Direction int32  `json:"direction"`
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

// route
type AddRouteForm struct {
	Name    string `json:"name"`
	Cidr    string `json:"cidr"`
	Nexthop string `json:"nexthop"`
}

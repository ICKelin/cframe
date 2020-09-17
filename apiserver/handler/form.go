package handler

type AddEdgeForm struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	HostAddr string `json:"hostaddr"`
	Cidr     string `json:"cidr"`
}

type DeleteEdgeForm struct {
	Name string `json:"name"`
}

type SignupForm struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SigninForm struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

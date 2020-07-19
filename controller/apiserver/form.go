package apiserver

type AddEdageForm struct {
	Name     string `json:"name"`
	HostAddr string `json:"hostaddr"`
	Cidr     string `json:"cidr"`
}

type DeleteEdageForm struct {
	Name string `json:"name"`
}

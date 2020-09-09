package main

type AddEdgeForm struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	HostAddr string `json:"hostaddr"`
	Cidr     string `json:"cidr"`
}

type DeleteEdgeForm struct {
	Name string `json:"name"`
}

type AddAccessForm struct {
	Type         string `json:"type"`
	AccessKey    string `json:"access_key"`
	AccessSecret string `json:"access_secret"`
}

type DeleteAccessForm struct {
	Type string `json:"type"`
}

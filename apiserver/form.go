package main

type AddEdgeForm struct {
	Name     string `json:"name"`
	HostAddr string `json:"hostaddr"`
	Cidr     string `json:"cidr"`
}

type DeleteEdgeForm struct {
	Name string `json:"name"`
}

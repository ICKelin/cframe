package controller

import "net/http"

func Main() {
	http.HandleFunc("/api/v1/nodes", getNodes)
	http.ListenAndServe(":10033", nil)
}

func getNodes(w http.ResponseWriter, r *http.Request) {
	// TODO:
}

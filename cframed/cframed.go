package cframed

func Main() {
	s := NewServer(":10222")
	s.ListenAndServe()
}

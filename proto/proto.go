package proto

type Handshake struct {
	IP   string `json:"ip"`
	Mask int32  `json:"mask"`
	// add other message
}

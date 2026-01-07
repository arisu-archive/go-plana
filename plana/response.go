package plana

type ResponseData struct {
	Protocol string `json:",omitempty,omitzero"`
	Packet   string `json:",omitempty,omitzero"`
}

package plana

import "github.com/arisu-archive/plana-protos/protos"

type ResponseData struct {
	Protocol protos.Protocol `json:",omitempty,omitzero"`
	Packet   string          `json:",omitempty,omitzero"`
}

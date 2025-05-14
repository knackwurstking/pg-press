package api

import (
	"testing"
)

func TestMicroRequestError_InvalidServer(t *testing.T) {
	d := &Device{
		Server: &Server{
			Addr: "1111:1111",
			Name: "Not existing Server",
		},
	}
	r := &MicroRequest{
		ID:      MicroIDNoResponse,
		Group:   MicroGroupConfig,
		Command: "not-existing-command",
	}
	r.Send(d)
	if d.Error == "" {
		t.Errorf("Error expected, because the server address is invalid, but go nothing")
	} else {
		t.Log(d.Error)
	}
}

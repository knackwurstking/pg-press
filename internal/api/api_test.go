package api

import (
	"testing"
)

func TestMicroRequestError_InvalidCommand(t *testing.T) {
	apiConfig, _ := GetConfig(".api.yaml")
	devices := []*Device{}
	for _, d := range apiConfig.Devices {
		devices = append(devices, d)
	}

	r := &MicroRequest{
		ID:      MicroIDDefault,
		Type:    MicroTypeGET,
		Group:   MicroGroupConfig,
		Command: "not-existing-command",
	}

	for _, d := range devices {
		respData, err := r.Send(d)
		if err != nil {
			continue
		}

		_, err = ParseMicroResponse[any](respData)
		if err == nil {
			t.Errorf("Error expected, because the command does not exists, but go nothing. server=%+v", d)
		} else {
			t.Log(err)
		}
	}
}

func TestMicroRequestError_InvalidGroup(t *testing.T) {
	apiConfig, _ := GetConfig(".api.yaml")
	devices := []*Device{}
	for _, d := range apiConfig.Devices {
		devices = append(devices, d)
	}

	r := &MicroRequest{
		ID:      MicroIDDefault,
		Type:    MicroTypeGET,
		Group:   "wrong-group",
		Command: "not-existing-command",
	}

	for _, d := range devices {
		respData, err := r.Send(d)
		if err != nil {
			continue
		}

		_, err = ParseMicroResponse[any](respData)
		if err == nil {
			t.Errorf("Error expected, because the command does not exists, but go nothing. server=%+v", d)
		} else {
			t.Log(err)
		}
	}
}

func TestMicroRequestError_InvalidType(t *testing.T) {
	apiConfig, _ := GetConfig(".api.yaml")
	devices := []*Device{}
	for _, d := range apiConfig.Devices {
		devices = append(devices, d)
	}

	r := &MicroRequest{
		ID:      MicroIDDefault,
		Type:    "wrong-type",
		Group:   "wrong-group",
		Command: "not-existing-command",
	}

	for _, d := range devices {
		respData, err := r.Send(d)
		if err != nil {
			continue
		}

		_, err = ParseMicroResponse[any](respData)
		if err == nil {
			t.Errorf("Error expected, because the command does not exists, but go nothing. server=%+v", d)
		} else {
			t.Log(err)
		}
	}
}

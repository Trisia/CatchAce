package main

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestSend(t *testing.T) {
	msg := Msg{Action: "OwnerExit"}
	bins, err := json.Marshal(msg)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(string(bins))
}

func TestPlayer_WaitMsg(t *testing.T) {
	reply := []byte(`{"Action":"DoDraw","Username":"Cliven"}`)
	var resp Msg
	err := json.Unmarshal(reply, &resp)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(resp)
}

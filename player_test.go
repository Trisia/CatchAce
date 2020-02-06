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

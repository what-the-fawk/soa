package e2e_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os/exec"
	"soa/common"
	"testing"
	"time"
)

func TestRegisterAuth(t *testing.T) {

	cmd := exec.Command("docker-compose", "up")
	go cmd.Run()

	time.Sleep(60 * time.Second)

	logpwd := &common.AuthInfo{
		Login:    "test1",
		Password: "pwd",
	}

	body, _ := json.Marshal(logpwd)
	req, err := http.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))

	if err != nil {
		t.Fatal(err.Error())
	}

}

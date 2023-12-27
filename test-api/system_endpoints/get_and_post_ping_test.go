package system_endpoints

import (
	"encoding/json"
	test_api "github.com/syncthing/syncthing/test-api"
	"log"
	"os"
	"os/exec"
	"reflect"
	"testing"
	"time"
)

type PingPong struct {
	Ping string `json:"ping"`
}

func Test_GetPing_ShouldReturn_PingPongObject(t *testing.T) {
	expected := "pong"
	result := CallPingEndpoint("GET", t)
	if !reflect.DeepEqual(result.Ping, expected) {
		t.Fatalf("Expected: %v, Got: %v", expected, result.Ping)
	}
}

func Test_PostPing_ShouldReturn_PingPongObject(t *testing.T) {
	expected := "pong"
	result := CallPingEndpoint("POST", t)
	if !reflect.DeepEqual(result.Ping, expected) {
		t.Fatalf("Expected: %v, Got: %v", expected, result.Ping)
	}
}

func CallPingEndpoint(method string, t *testing.T) PingPong {
	binPath := "../../bin"
	homePath := "../api-test-home"

	address, apikey, err := test_api.GetAddressAndApiKey(binPath, homePath)
	if err != nil {
		t.Fatalf("Could not get address and apikey: %v", err)
	}

	cmd := exec.Command(binPath+"/syncthing", "--no-browser", "--home", homePath)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		t.Fatal("could not start syncthing process")
	}

	defer func() {
		if err := cmd.Process.Kill(); err != nil {
			log.Fatalf("Warning: Error killing Syncthing process: %v", err)
		}
	}()

	baseURL := "http://" + address

	timeout := time.After(30 * time.Second)
	tick := time.Tick(1 * time.Second)

	for {
		select {
		case <-timeout:
			t.Fatal("Syncthing startup took to long")
		case <-tick:
			if test_api.CheckServerHealth(baseURL) {
				t.Log("Syncthing is running..")
				goto SyncthingReady
			}
		}
	}

SyncthingReady:
	pingUrl := "http://" + address + "/rest/system/ping"
	response, err := test_api.MakeHttpRequest(method, apikey, pingUrl)
	if err != nil {
		t.Fatalf("Could not make http request: %v", err)
	}

	var pingPong PingPong
	if err := json.NewDecoder(response.Body).Decode(&pingPong); err != nil {
		t.Fatalf("Could not decode json response: %v", err)
	}
	return pingPong
}

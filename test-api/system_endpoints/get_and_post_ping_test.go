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

	// Get address and apikey from running syncthing instance.
	address, apikey, err := test_api.GetAddressAndApiKey(binPath, homePath)
	if err != nil {
		t.Fatalf("Could not get address and apikey: %v", err)
	}

	// Get a cmd struct to execute syncthing from.
	cmd := exec.Command(binPath+"/syncthing", "--no-browser", "--home", homePath)

	// We want the printing of messages and errors to be the same as the operative system.
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		t.Fatal("could not start syncthing process")
	}

	// Defer the shutting down of syncthing instance to occur last in this function.
	defer func() {
		if err := cmd.Process.Kill(); err != nil {
			log.Fatalf("Warning: Error killing Syncthing process: %v", err)
		}
	}()

	// Setup REST API url to call
	baseURL := "http://" + address

	// Timeout set for when we stop checking if syncthing has started.
	timeout := time.After(30 * time.Second)
	// Tick set for the interval of checking if syncthing is up and running
	tick := time.Tick(1 * time.Second)

	for {
		select {
		// If timeout has passed, syncthing has not started correctly => fail the test.
		case <-timeout:
			t.Fatal("Syncthing startup took to long")
		// If timeout has not passed, check if syncthing is running for each tick.
		case <-tick:
			if test_api.CheckServerHealth(baseURL) {
				t.Log("Syncthing is running..")
				// Label for the actual test. => Start the API testing logic.
				goto SyncthingReady
			}
		}
	}

	// The actual testing logic
SyncthingReady:
	errorURL := "http://" + address + "/rest/system/ping"
	response, err := test_api.MakeHttpRequest(method, apikey, errorURL)
	if err != nil {
		t.Fatalf("Could not make http request: %v", err)
	}

	var pingPong PingPong
	if err := json.NewDecoder(response.Body).Decode(&pingPong); err != nil {
		t.Fatalf("Could not decode json response: %v", err)
	}
	return pingPong
}

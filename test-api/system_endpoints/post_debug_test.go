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

func Test_PostDebug_ShouldEnableAndDisable_DebugFacilities(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected DebugInfo
	}{
		"one-value": {input: "config", expected: DebugInfo{
			[]string{"config"},
			DefaultDebugInfo.Facilities,
		}},
		"multiple-values": {input: "config,db,sha256", expected: DebugInfo{
			[]string{"config", "db", "sha256"},
			DefaultDebugInfo.Facilities,
		}},
	}

	binPath := "../../bin"
	homePath := "../api-test-home"

	// Get address and apikey from running syncthing instance.
	address, apikey, err := test_api.GetAddressAndApiKey(binPath, homePath)
	if err != nil {
		t.Fatalf("Could not get address and apikey: %s", err)
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
	for name, testCase := range tests {
		t.Run(name, func(t *testing.T) {
			debugPostURL := "http://" + address + "/rest/system/debug/?enable=" + testCase.input
			debugGetURL := "http://" + address + "/rest/system/debug/"

			_, err := test_api.MakeHttpRequest("POST", apikey, debugPostURL)
			if err != nil {
				t.Fatalf("Failed to browse: %s", err)
			}

			response, err := test_api.MakeHttpRequest("GET", apikey, debugGetURL)
			if err != nil {
				t.Fatalf("Failed to browse: %s", err)
			}

			var resultDebugInfo DebugInfo
			if err := json.NewDecoder(response.Body).Decode(&resultDebugInfo); err != nil {
				t.Fatalf("Failed to decode response: %s", err)
			}

			if !reflect.DeepEqual(resultDebugInfo, testCase.expected) {
				t.Errorf("Expected %s, got %s", testCase.expected, resultDebugInfo)
			}

			cleanupURL := "http://" + address + "/rest/system/debug/?disable=" + testCase.input
			_, err = test_api.MakeHttpRequest("POST", apikey, cleanupURL)
			if err != nil {
				t.Fatalf("Failed to browse: %s", err)
			}

			cleanupResponse, err := test_api.MakeHttpRequest("GET", apikey, debugGetURL)
			if err != nil {
				t.Fatalf("Failed to browse: %s", err)
			}

			var cleanedUpDebugInfo DebugInfo
			if err := json.NewDecoder(cleanupResponse.Body).Decode(&cleanedUpDebugInfo); err != nil {
				t.Fatalf("Failed to decode response: %s", err)
			}

			if !reflect.DeepEqual(cleanedUpDebugInfo, DefaultDebugInfo) {
				t.Errorf("Expected %s, got %s", DefaultDebugInfo, cleanedUpDebugInfo)
			}

		})
	}

}

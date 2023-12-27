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
	homePath := "../api-test-home/debug"

	address, apikey, err := test_api.GetAddressAndApiKey(binPath, homePath)
	if err != nil {
		t.Fatalf("Could not get address and apikey: %s", err)
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

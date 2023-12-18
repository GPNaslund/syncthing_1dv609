package system_endpoints

import (
	"encoding/json"
	test_api "github.com/syncthing/syncthing/test-api"
	"reflect"
	"testing"
	"time"
)

import (
	"log"
	"os"
	"os/exec"
)

type DebugInfo struct {
	Enabled    []string          `json:"enabled"`
	Facilities map[string]string `json:"facilities"`
}

var DefaultDebugInfo = DebugInfo{
	Enabled: []string{}, // Assuming no facilities are enabled in the expected response
	Facilities: map[string]string{
		"api":             "REST API",
		"app":             "Main run facility",
		"backend":         "The database backend",
		"beacon":          "Multicast and broadcast discovery",
		"config":          "Configuration loading and saving",
		"connections":     "Connection handling",
		"db":              "The database layer",
		"dialer":          "Dialing connections",
		"discover":        "Remote device discovery",
		"events":          "Event generation and logging",
		"fs":              "Filesystem access",
		"main":            "Main package",
		"model":           "The root hub",
		"nat":             "NAT discovery and port mapping",
		"pmp":             "NAT-PMP discovery and port mapping",
		"protocol":        "The BEP protocol",
		"relay":           "",
		"scanner":         "File change detection and hashing",
		"sha256":          "SHA256 hashing package",
		"stats":           "Persistent device and folder statistics",
		"stun":            "STUN functionality",
		"sync":            "Mutexes",
		"upgrade":         "Binary upgrades",
		"upnp":            "UPnP discovery and port mapping",
		"ur":              "Usage reporting",
		"versioner":       "File versioning",
		"walkfs":          "Filesystem access while walking",
		"watchaggregator": "Filesystem event watcher",
	},
}

func Test_GetDebug_ShouldReturn_DebugInformation(t *testing.T) {
	// Setup path to bin and home
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
	// Construct an url for the REST API, sets current=testdata to start fetching
	// directory information from the folder testdata
	debugURL := "http://" + address + "/rest/system/debug/"

	// Get http response.
	response, err := test_api.MakeHttpRequest("GET", apikey, debugURL)
	if err != nil {
		t.Fatalf("Failed to browse: %s", err)
	}

	var resultDebugInfo DebugInfo
	if err := json.NewDecoder(response.Body).Decode(&resultDebugInfo); err != nil {
		t.Fatalf("Failed to decode response: %s", err)
	}

	// Check if expected and result are equal.
	if !reflect.DeepEqual(resultDebugInfo, DefaultDebugInfo) {
		t.Errorf("Expected %s, got %s", DefaultDebugInfo, resultDebugInfo)
	}
}

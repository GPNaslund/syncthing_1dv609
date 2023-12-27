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
	debugURL := "http://" + address + "/rest/system/debug/"

	response, err := test_api.MakeHttpRequest("GET", apikey, debugURL)
	if err != nil {
		t.Fatalf("Failed to browse: %s", err)
	}

	var resultDebugInfo DebugInfo
	if err := json.NewDecoder(response.Body).Decode(&resultDebugInfo); err != nil {
		t.Fatalf("Failed to decode response: %s", err)
	}

	if !reflect.DeepEqual(resultDebugInfo, DefaultDebugInfo) {
		t.Errorf("Expected %s, got %s", DefaultDebugInfo, resultDebugInfo)
	}
}

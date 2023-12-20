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

type PathData struct {
	BaseDirConfig   string `json:"baseDir-config"`
	BaseDirData     string `json:"baseDir-data"`
	BaseDirUserHome string `json:"baseDir-userHome"`
	CertFile        string `json:"certFile"`
	Config          string `json:"config"`
	Database        string `json:"database"`
	DefFolder       string `json:"defFolder"`
	HttpsCertFile   string `json:"httpsCertFile"`
	HttpsKeyFile    string `json:"httpsKeyFile"`
	LogFile         string `json:"logFile"`
}

func Test_GetPaths_ShouldReturn_ConfigPaths(t *testing.T) {
	allPathData := MakeGetCallToPathsEndpoint(t)

	validatePath := func(path string) {
		if path != "-" && path != "" {
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Fatalf("Path does not exists: %v", path)
			} else if err != nil {
				t.Fatalf("Error checking path: %v, error: %v", path, err)
			}
		}
	}

	values := reflect.ValueOf(allPathData)
	for i := 0; i < values.NumField(); i++ {
		path := values.Field(i).String()
		validatePath(path)
	}
}

func MakeGetCallToPathsEndpoint(t *testing.T) PathData {
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
	logURL := "http://" + address + "/rest/system/paths"
	response, err := test_api.MakeHttpRequest("GET", apikey, logURL)
	if err != nil {
		t.Fatalf("Could not do post request: %v", err)
	}
	var pathData PathData
	if err := json.NewDecoder(response.Body).Decode(&pathData); err != nil {
		t.Fatalf("Could not decode JSON: %v", err)
	}
	return pathData
}

package system_endpoints

import (
	"bytes"
	test_api "github.com/syncthing/syncthing/test-api"
	"log"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"testing"
	"time"
)

func Test_PostError_ShouldAddANewErrorMessage_ToSystem(t *testing.T) {
	newErrorMessage := "Error message for testing api"
	AddNewErrorMessage_ThroughEndpoint(t, newErrorMessage)
	logfileErrors, err := ParseLogFileForErrors("../api-test-home/syncthing.log")
	if err != nil {
		t.Fatalf("Could not parse log file: %v", err)
	}
	lastLogFileEntry := logfileErrors[len(logfileErrors)-1]
	if !reflect.DeepEqual(lastLogFileEntry, newErrorMessage) {
		t.Fatalf("Expected: %v, Got: %v", newErrorMessage, lastLogFileEntry)
	}
}

func AddNewErrorMessage_ThroughEndpoint(t *testing.T, newErrorMessage string) {
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

	// The actual testing logic
SyncthingReady:
	errorURL := "http://" + address + "/rest/system/error/"
	client := &http.Client{}

	requestBody := bytes.NewReader([]byte(newErrorMessage))
	request, err := http.NewRequest("POST", errorURL, requestBody)
	if err != nil {
		t.Fatalf("Could not create post request: %v", err)
	}
	request.Header.Set("X-Api-Key", apikey)
	_, err = client.Do(request)
	if err != nil {
		t.Fatalf("Could not do post request: %v", err)
	}
}

package system_endpoints

import (
	test_api "github.com/syncthing/syncthing/test-api"
	"log"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"
)

func Test_ErrorClear_ShouldClearLogFromErrors(t *testing.T) {
	AddNewErrorMessage_ThroughEndpoint(t, "Error message for test clear endpoint")
	listOfLogErrorsBefore, err := ParseLogFileForErrors("../api-test-home/syncthing.log")
	if err != nil {
		t.Fatalf("Could not parse log file for errors: %v", err)
	}
	Call_ErrorClear_Endpoint(t)
	listOfLogErrorsAfter, err := ParseLogFileForErrors("../api-test-home/syncthing.log")
	if err != nil {
		t.Fatalf("Could not parse log file for errors: %v", err)
	}

	amountOfLogErrorsBeforeClearing := len(listOfLogErrorsBefore)
	amountOfLogErrorsAfterClearing := len(listOfLogErrorsAfter)

	if amountOfLogErrorsBeforeClearing >= amountOfLogErrorsAfterClearing {
		t.Fatalf("Logs after clearing: %v, Logs before clearing: %v. No errors got cleared!",
			amountOfLogErrorsAfterClearing, amountOfLogErrorsBeforeClearing)
	}
}

func Call_ErrorClear_Endpoint(t *testing.T) {
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
	errorURL := "http://" + address + "/rest/system/error/clear"
	client := &http.Client{}

	request, err := http.NewRequest("POST", errorURL, nil)
	if err != nil {
		t.Fatalf("Could not create post request: %v", err)
	}
	request.Header.Set("X-Api-Key", apikey)
	_, err = client.Do(request)
	if err != nil {
		t.Fatalf("Could not do post request: %v", err)
	}
}

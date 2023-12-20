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

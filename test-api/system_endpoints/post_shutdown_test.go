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

func Test_PostShutdown_ShouldCloseSyncthingInstance(t *testing.T) {
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
	shutdownURL := "http://" + address + "/rest/system/shutdown/"
	// Command line execution of "syncthing --no-browser --home=api-test-home" whilst
	// being inside the syncthing bin folder. exec.Command starts a process/program.
	cmd = exec.Command(binPath+"/syncthing", "--no-browser", "--home", homePath)

	// We want the printing of messages and errors to be the same as the operative system.
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()

	// Get a HTTP Client to make calls from.
	client := &http.Client{}

	// Create a new HTTP request to send from the HTTP client.
	request, err := http.NewRequest("POST", shutdownURL, nil)
	if err != nil {
		t.Fatalf("Could not create new http request: %v", err)
	}

	// Sets the API key
	request.Header.Set("X-API-Key", apikey)

	_, err = client.Do(request)
	if err != nil {
		t.Fatalf("Could not do request: %v", err)
	}

	// Create a new HTTP request to send from the HTTP client.
	request, err = http.NewRequest("POST", shutdownURL, nil)
	if err != nil {
		t.Fatalf("Could not create new http request: %v", err)
	}

	// Sets the API key
	request.Header.Set("X-API-Key", apikey)

	response, err := client.Do(request)
	if response != nil {
		t.Fatalf("Syncthing did not quit!")
	}

}

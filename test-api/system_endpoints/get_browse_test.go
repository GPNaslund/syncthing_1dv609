package system_endpoints

import (
	"encoding/json"
	"github.com/syncthing/syncthing/test-api"
	"log"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"
)

func Test_GetBrowse_ShouldReturn_ListOfDirectories(t *testing.T) {
	// Setup path to bin and home
	binPath := "../../bin"
	homePath := "../api-test-home"

	// Get address and apikey from running syncthing instance.
	address, apikey, err := api_test.GetAddressAndApiKey(binPath, homePath)
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
	healthCheckUrl := "http://" + address

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
			if api_test.CheckServerHealth(healthCheckUrl) {
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
	url := "http://" + address + "/rest/system/browse?current=testdata/"

	// Get a list of the directories found by the REST API.
	resultDirectories, err := get_browse(apikey, url)
	if err != nil {
		t.Fatalf("Failed to browse: %s", err)
	}

	// Corresponds to the folder structure inside system_endpoints/testdata/
	expectedDirectories := []string{
		"testdata/folder1",
		"testdata/folder2",
		"testdata/folder3",
		"testdata/folder4",
		"testdata/folder5",
	}

	// Normalize paths in both expected and actual results
	normalizedExpectedDirectories := make([]string, len(expectedDirectories))
	for i, dir := range expectedDirectories {
		normalizedExpectedDirectories[i] = strings.ReplaceAll(dir, "\\", "/")
	}

	normalizedResultDirectories := make([]string, len(resultDirectories))
	for i, dir := range resultDirectories {
		normalizedResultDirectories[i] = strings.ReplaceAll(dir, "\\", "/")
	}

	// Check if expected and result are equal.
	if !reflect.DeepEqual(normalizedResultDirectories, normalizedExpectedDirectories) {
		t.Errorf("Expected %s, got %s", normalizedExpectedDirectories, normalizedResultDirectories)
	}
}

// Function for calling GET on the provided url and returning the result.
func get_browse(apiKey, url string) ([]string, error) {
	// Get a HTTP Client to make calls from.
	client := &http.Client{}

	// Create a new HTTP request to send from the HTTP client.
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Sets the API key
	request.Header.Set("X-API-Key", apiKey)

	// Perform the GET request to provided url with provided api key.
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	// Close the response body when the function is done.
	defer response.Body.Close()

	var resultDirectories []string
	// Decode the json response body
	responseDecoded := json.NewDecoder(response.Body).Decode(&resultDirectories)
	if responseDecoded != nil {
		return nil, err
	}
	return resultDirectories, nil
}

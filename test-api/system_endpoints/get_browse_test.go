package system_endpoints

import (
	"encoding/json"
	"github.com/syncthing/syncthing/test-api"
	"log"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"
)

func Test_GetBrowse_ShouldReturn_ListOfDirectories(t *testing.T) {
	binPath := "../../bin"
	homePath := "../api-test-home/browse"

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
	browseUrl := "http://" + address + "/rest/system/browse?current=testdata/"

	response, err := test_api.MakeHttpRequest("GET", apikey, browseUrl)
	if err != nil {
		t.Fatalf("Failed to browse: %s", err)
	}

	defer response.Body.Close()

	var resultDirectories []string

	if err := json.NewDecoder(response.Body).Decode(&resultDirectories); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Corresponds to the folder structure inside system_endpoints/testdata/
	expectedDirectories := []string{
		"testdata/folder1",
		"testdata/folder2",
		"testdata/folder3",
		"testdata/folder4",
		"testdata/folder5",
	}

	normalizedExpectedDirectories := make([]string, len(expectedDirectories))
	for i, dir := range expectedDirectories {
		normalizedExpectedDirectories[i] = strings.ReplaceAll(dir, "\\", "/")
	}

	normalizedResultDirectories := make([]string, len(resultDirectories))
	for i, dir := range resultDirectories {
		normalizedResultDirectories[i] = strings.ReplaceAll(dir, "\\", "/")
	}

	if !reflect.DeepEqual(normalizedResultDirectories, normalizedExpectedDirectories) {
		t.Errorf("Expected %s, got %s", normalizedExpectedDirectories, normalizedResultDirectories)
	}
}

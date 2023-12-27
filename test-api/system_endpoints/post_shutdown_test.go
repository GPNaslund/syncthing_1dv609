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
	shutdownURL := "http://" + address + "/rest/system/shutdown/"

	cmd = exec.Command(binPath+"/syncthing", "--no-browser", "--home", homePath)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()

	client := &http.Client{}

	request, err := http.NewRequest("POST", shutdownURL, nil)
	if err != nil {
		t.Fatalf("Could not create new http request: %v", err)
	}

	request.Header.Set("X-API-Key", apikey)

	_, err = client.Do(request)
	if err != nil {
		t.Fatalf("Could not do request: %v", err)
	}

	request, err = http.NewRequest("POST", shutdownURL, nil)
	if err != nil {
		t.Fatalf("Could not create new http request: %v", err)
	}

	request.Header.Set("X-API-Key", apikey)

	response, err := client.Do(request)
	if response != nil {
		t.Fatalf("Syncthing did not quit!")
	}

}

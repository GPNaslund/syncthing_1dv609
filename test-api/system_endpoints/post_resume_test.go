package system_endpoints

import (
	"fmt"
	apitest "github.com/syncthing/syncthing/test-api"
	"log"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"
)

func Test_PostResume_ShouldReturn_StatusCode200(t *testing.T) {
	binPath := "../../bin"
	homePath := "../get-connections-test-home"

	cmd := exec.Command(binPath+"/syncthing", "--no-browser", "--home", homePath)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		t.Fatal("could not start syncthing process")
	}

	address, apikey, err2 := apitest.GetAddressAndApiKey(binPath, homePath)
	if err2 != nil {
		t.Fatalf("Could not get address and apikey: %s", err2)
	}

	defer func() {
		if err := cmd.Process.Kill(); err != nil {
			log.Fatalf("Warning: Error killing Syncthing process: %v", err)
		}
	}()

	healthCheckUrl := "http://" + address

	timeout := time.After(30 * time.Second)
	tick := time.Tick(1 * time.Second)

	for {
		select {
		case <-timeout:
			t.Fatal("Syncthing startup took to long")
		case <-tick:
			if apitest.CheckServerHealth(healthCheckUrl) {
				t.Log("Syncthing is running..")
				goto SyncthingReady
			}
		}
	}

SyncthingReady:
	url := "http://" + address + "/rest/system/pause"
	resp, err := apitest.MakeHttpRequest("POST", apikey, url)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Failed to retrieve configuration. Status code:", resp.StatusCode)
		return
	}

	time.Sleep(5 * time.Second)

	url = "http://" + address + "/rest/system/resume"
	resp, err = apitest.MakeHttpRequest("POST", apikey, url)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Failed to resume devices")
	}

}

package system_endpoints

import (
	"encoding/json"
	"fmt"
	apitest "github.com/syncthing/syncthing/test-api"
	"log"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"
)

func Test_GetDiscovery_ShouldReturn_DiscoveryCache(t *testing.T) {

	// Setup path to bin and home
	binPath := "../../bin"
	homePath := "../get-discovery-test-home"

	// Get a cmd struct to execute syncthing from.
	cmd := exec.Command(binPath+"/syncthing", "--no-browser", "--home", homePath)

	// We want the printing of messages and errors to be the same as the operative system.
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		t.Fatal("could not start syncthing process")
	}

	// Get address and apikey from running syncthing instance.
	address, apikey, err2 := apitest.GetAddressAndApiKey(binPath, homePath)
	if err2 != nil {
		t.Fatalf("Could not get address and apikey: %s", err2)
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
			if apitest.CheckServerHealth(healthCheckUrl) {
				t.Log("Syncthing is running..")
				// Label for the actual test. => Start the API testing logic.
				goto SyncthingReady
			}
		}
	}

SyncthingReady:
	//Device ID and name already present in test home folder
	//deviceID := "H67OXGJ-BSITBYE-MZ3BJPH-6BMIGIE-7PROEHT-6QYVQVI-C7INUEY-LPP6UQP"

	devices := [3]string{
		"6VKB3M2-EN7G6J6-5SHPI3Y-GEWMIAJ-DJJSQLY-CKJFXQ3-DMUGVK7-X46RAQU",
		"FMXYAGO-SRWPBMB-5G6VEKW-WRSXRXM-ENB4S6H-KVPHQUP-4BQE33E-OKOWYAL",
		"XQF4EQQ-RDZXUJT-NTZILHT-VOJTZCE-A2OLLBX-L6YUODL-UVC6AXK-KSPDJAQ",
	}

	//Get devices configured with the current instance of syncthing
	url := "http://" + address + "/rest/system/discovery"
	resp, err := apitest.MakeHttpRequest("GET", apikey, url)
	defer resp.Body.Close()

	//Check response code of response from syncthing
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Failed to retrieve configuration. Status code:", resp.StatusCode)
		return
	}

	//Store the devices received in the response
	var config map[string]map[string]interface{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}

	for i := 0; i < len(devices); i++ {
		if config[devices[i]] == nil {
			t.Errorf("Device: %s was not found", devices[i])
		}
	}

}

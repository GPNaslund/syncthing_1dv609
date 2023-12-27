package system_endpoints

/*
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
	binPath := "../../bin"
	homePath := "../get-discovery-test-home"

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
	//Device ID and name already present in test home folder
	//deviceID := "H67OXGJ-BSITBYE-MZ3BJPH-6BMIGIE-7PROEHT-6QYVQVI-C7INUEY-LPP6UQP"

	devices := [3]string{
		"6VKB3M2-EN7G6J6-5SHPI3Y-GEWMIAJ-DJJSQLY-CKJFXQ3-DMUGVK7-X46RAQU",
		"FMXYAGO-SRWPBMB-5G6VEKW-WRSXRXM-ENB4S6H-KVPHQUP-4BQE33E-OKOWYAL",
		"XQF4EQQ-RDZXUJT-NTZILHT-VOJTZCE-A2OLLBX-L6YUODL-UVC6AXK-KSPDJAQ",
	}

	url := "http://" + address + "/rest/system/discovery"
	resp, err := apitest.MakeHttpRequest("GET", apikey, url)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Failed to retrieve configuration. Status code:", resp.StatusCode)
		return
	}

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
*/

package system_endpoints

import (
	"encoding/json"
	"fmt"
	apitest "github.com/syncthing/syncthing/test-api"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"
)

type Device struct {
	Id   string `json:"deviceID"`
	Name string `json:"Name"`
}

func Test_GetConfigDevices_ShouldReturn_ListOfDevices(t *testing.T) {
	binPath := "../../bin"
	homePath := "../get-config-devices-test-home"

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
	deviceID := "H67OXGJ-BSITBYE-MZ3BJPH-6BMIGIE-7PROEHT-6QYVQVI-C7INUEY-LPP6UQP"
	deviceName := "Phone"

	url := "http://" + address + "/rest/config/devices/"
	resp, err := apitest.MakeHttpRequest("GET", apikey, url)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Failed to retrieve configuration. Status code:", resp.StatusCode)
		return
	}

	byteValue, err := io.ReadAll(resp.Body)
	var devices []Device

	err = json.Unmarshal(byteValue, &devices)
	if err != nil {
		panic(err)
	}

	if !(devices[0].Id == deviceID && devices[0].Name == deviceName) {
		t.Errorf("Expected ID:%s,Name:%s got ID:%s,Name:%s", deviceID, deviceName, devices[1].Id, devices[1].Name)
	}

}

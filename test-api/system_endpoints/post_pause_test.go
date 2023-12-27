package system_endpoints

import (
	"encoding/xml"
	test_api "github.com/syncthing/syncthing/test-api"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"
)

type Config struct {
	XMLName xml.Name       `xml:"configuration"`
	Folders []ConfigFolder `xml:"folder"`
	Devices []ConfigDevice `xml:"device"`
}

type ConfigFolder struct {
	Devices []ConfigDevice `xml:"device"`
}

type ConfigDevice struct {
	ID string `xml:"id,attr"`
}

func Test_PostPause_ShouldSuccessfully_PauseAllDevices(t *testing.T) {
	response := MakePostRequestToPauseEndpoint("", t)
	defer response.Body.Close()

	if response.StatusCode != 200 {
		t.Fatalf("Pause response was not 200: %v", response.StatusCode)
	}
}

func Test_PostPause_WithDeviceId_ShouldSuccesfully_Pause(t *testing.T) {
	configFile, err := os.Open("../api-test-home/config.xml")
	if err != nil {
		t.Fatalf("Could not open config file: %v", err)
	}

	defer func() {
		if err := configFile.Close(); err != nil {
			log.Printf("Warning: Error closing config file: %v", err)
		}
	}()

	byteValue, err := io.ReadAll(configFile)
	if err != nil {
		t.Fatalf("Could not read config file: %v", err)
	}

	var config Config
	err = xml.Unmarshal(byteValue, &config)

	if err != nil {
		t.Fatalf("Could not unmarshal XML file: %v", err)
	}

	var deviceID string
	if len(config.Devices) > 0 {
		deviceID = config.Devices[0].ID
	} else if len(config.Folders) > 0 && len(config.Folders[0].Devices) > 0 {
		// If no device at root level, check within folders
		deviceID = config.Folders[0].Devices[0].ID
	}

	if deviceID == "" {
		t.Fatalf("No device ID found in config")
	}

	response := MakePostRequestToPauseEndpoint("?device="+deviceID, t)

	defer response.Body.Close()

	if response.StatusCode != 200 {
		t.Fatalf("Pause response was not 200: %v", response.StatusCode)
	}
}

func MakePostRequestToPauseEndpoint(deviceIdParam string, t *testing.T) *http.Response {
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
	errorURL := "http://" + address + "/rest/system/pause" + deviceIdParam
	client := &http.Client{}

	request, err := http.NewRequest("POST", errorURL, nil)
	if err != nil {
		t.Fatalf("Could not create post request: %v", err)
	}
	request.Header.Set("X-Api-Key", apikey)
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("Could not do post request: %v", err)
	}

	return response
}

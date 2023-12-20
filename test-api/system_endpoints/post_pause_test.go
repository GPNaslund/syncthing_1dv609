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
	// Try to open the xml config file
	configFile, err := os.Open("../api-test-home/config.xml")
	if err != nil {
		t.Fatalf("Could not open config file: %v", err)
	}

	// Before our method returns, close the xml file.
	defer func() {
		if err := configFile.Close(); err != nil {
			log.Printf("Warning: Error closing config file: %v", err)
		}
	}()

	// Read the contents of the config
	byteValue, err := io.ReadAll(configFile)
	if err != nil {
		t.Fatalf("Could not read config file: %v", err)
	}

	// Creates a Configuration (the struct declared at the top)
	// variable. Unmarshal will "unpack" the contents of the XML
	// into the corresponding fields of the struct. Since we specified
	// address and api key in the struct (declared at the top), that is
	// all we get.
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

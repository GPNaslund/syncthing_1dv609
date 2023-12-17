package api_test

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
)

// Configuration Struct/"Object" for parsing config XML.
type Configuration struct {
	// Only a field for the GUI part of the config
	// that holds the API key and Address.
	GUI GUI `xml:"gui"`
}

// GUI I only want the Address and APIKey from the GUI
// part of the config.
type GUI struct {
	Address string `xml:"address"`
	APIKey  string `xml:"apikey"`
}

func GetAddressAndApiKey(binPath string, homePath string) (string, string, error) {
	// Check if bin folder exists aka is syncthing compiled?
	_, err := os.Stat(binPath)
	if os.IsNotExist(err) {
		return "", "", errors.New("syncthing must be built to run the tests")
	}

	// Does the home folder exist? If not, run the helper method that generates it!
	_, err = os.Stat(homePath)
	if os.IsNotExist(err) {
		FirstTimeInitialization(binPath, homePath)
	}

	// Command line execution of "syncthing --no-browser --home=api-test-home" whilst
	// being inside the syncthing bin folder. exec.Command starts a process/program.
	cmd := exec.Command(binPath+"/syncthing", "--no-browser", "--home", homePath)

	// We want the printing of messages and errors to be the same as the operative system.
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		return "", "", errors.New("could not start syncthing process")
	}

	// Defer is used to execute something right before the method returns. We
	// can declare it anywhere, and it will still be executed in the end.
	// Here we call cmd.Process.Kill() to quit the running syncthing.
	defer func() {
		if err := cmd.Process.Kill(); err != nil {
			log.Fatalf("Warning: Error killing Syncthing process: %v", err)
		}
	}()

	// Try to open the xml config file
	configFile, err := os.Open(homePath + "/config.xml")
	if err != nil {
		return "", "", errors.New("could not open config file")
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
		return "", "", errors.New("could not read config file")
	}

	// Creates a Configuration (the struct declared at the top)
	// variable. Unmarshal will "unpack" the contents of the XML
	// into the corresponding fields of the struct. Since we specified
	// address and api key in the struct (declared at the top), that is
	// all we get.
	var config Configuration
	err = xml.Unmarshal(byteValue, &config)

	if err != nil {
		return "", "", errors.New("could not unmarshal xml file")
	}

	return config.GUI.Address, config.GUI.APIKey, nil
}

func FirstTimeInitialization(binPath string, homePath string) {
	// Generate necessary files like config, db etc. and
	// then close syncthing. Perfect for just getting a default config
	// that can be used for starting a syncthing instance.
	cmd := exec.Command(binPath+"/syncthing", "--generate", homePath, "--no-browser")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Failed to do first time initializing of Syncthing: %s", err)
	}
}

// Synchting has a REST API endpoint that does not need credentials to check
// if its healthy aka working. Is used to check if the spawned instance of
// syncthing is up and running, ready for REST API requests.

func CheckServerHealth(url string) bool {
	// Make a http request to the health REST API
	response, err := http.Get(url + "/rest/noauth/health")
	if err != nil {
		return false
	}
	// Close the response body in the end of this function.
	defer response.Body.Close()

	var result map[string]string
	responseContent := json.NewDecoder(response.Body).Decode(&result)
	if responseContent != nil {
		return false
	}

	status, ok := result["status"]
	return ok && status == "OK"
}

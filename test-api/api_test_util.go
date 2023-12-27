package test_api

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

type Configuration struct {
	// Only a field for the GUI part of the config
	// that holds the API key and Address.
	GUI GUI `xml:"gui"`
}

type GUI struct {
	Address string `xml:"address"`
	APIKey  string `xml:"apikey"`
}

func GetAddressAndApiKey(binPath string, homePath string) (string, string, error) {
	_, err := os.Stat(binPath)
	if os.IsNotExist(err) {
		return "", "", errors.New("syncthing must be built to run the tests")
	}

	_, err = os.Stat(homePath)
	if os.IsNotExist(err) {
		FirstTimeInitialization(binPath, homePath)
	}

	cmd := exec.Command(binPath+"/syncthing", "--no-browser", "--home", homePath)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		return "", "", errors.New("could not start syncthing process")
	}

	defer func() {
		if err := cmd.Process.Kill(); err != nil {
			log.Fatalf("Warning: Error killing Syncthing process: %v", err)
		}
	}()

	configFile, err := os.Open(homePath + "/config.xml")
	if err != nil {
		return "", "", errors.New("could not open config file")
	}

	defer func() {
		if err := configFile.Close(); err != nil {
			log.Printf("Warning: Error closing config file: %v", err)
		}
	}()

	byteValue, err := io.ReadAll(configFile)
	if err != nil {
		return "", "", errors.New("could not read config file")
	}

	var config Configuration
	err = xml.Unmarshal(byteValue, &config)

	if err != nil {
		return "", "", errors.New("could not unmarshal xml file")
	}

	return config.GUI.Address, config.GUI.APIKey, nil
}

func FirstTimeInitialization(binPath string, homePath string) {
	cmd := exec.Command(binPath+"/syncthing", "--generate", homePath, "--no-browser")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Failed to do first time initializing of Syncthing: %s", err)
	}
}

func CheckServerHealth(url string) bool {
	response, err := http.Get(url + "/rest/noauth/health")
	if err != nil {
		return false
	}

	defer response.Body.Close()

	var result map[string]string
	responseContent := json.NewDecoder(response.Body).Decode(&result)
	if responseContent != nil {
		return false
	}

	status, ok := result["status"]
	return ok && status == "OK"
}

func MakeHttpRequest(method, apiKey, url string) (*http.Response, error) {
	client := &http.Client{}

	request, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("X-API-Key", apiKey)

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

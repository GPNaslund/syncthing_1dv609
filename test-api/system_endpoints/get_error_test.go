package system_endpoints

import (
	"bufio"
	"encoding/json"
	test_api "github.com/syncthing/syncthing/test-api"
	"log"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"testing"
	"time"
)

type ListOfErrors struct {
	ErrorInfo []ErrorInfo `json:"errors"`
}

type ErrorInfo struct {
	When    string `json:"when"`
	Message string `json:"message"`
}

func Test_GetError_ShouldReturnSystemListOfErrors(t *testing.T) {
	var apiListErrors = GetErrorJsonData_From_GetError_Endpoint(t)
	logfileErrors, err := ParseLogFileForErrors("../api-test-home/syncthing.log")
	if err != nil {
		t.Fatalf("Could not parse log file: %v", err)
	}

	numOfApiErrors := len(apiListErrors.ErrorInfo)
	lastFileLogEntries := logfileErrors
	if len(logfileErrors) > numOfApiErrors {
		lastFileLogEntries = logfileErrors[len(logfileErrors)-numOfApiErrors:]
	}

	for i := 0; i < numOfApiErrors; i++ {
		if !reflect.DeepEqual(lastFileLogEntries[i], apiListErrors.ErrorInfo[i].Message) {
			t.Fatalf("Expected: %v, Got: %v", lastFileLogEntries[i], apiListErrors.ErrorInfo[i])
		}
	}
}

func GetErrorJsonData_From_GetError_Endpoint(t *testing.T) ListOfErrors {
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
	errorURL := "http://" + address + "/rest/system/error/"
	response, err := test_api.MakeHttpRequest("GET", apikey, errorURL)
	if err != nil {
		t.Fatalf("Could not make http request: %v", err)
	}

	var allErrors ListOfErrors
	if err := json.NewDecoder(response.Body).Decode(&allErrors); err != nil {
		t.Fatalf("Could not decode json response: %v", err)
	}
	return allErrors
}

func ParseLogFileForErrors(pathToLogFile string) ([]string, error) {
	file, err := os.Open(pathToLogFile)
	if err != nil {
		log.Fatalf("Could not open log file: %v", err)
		return nil, err
	}
	defer file.Close()

	errorMessages := []string{}

	re := regexp.MustCompile(`WARNING: (.*)`)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindStringSubmatch(line)
		if matches != nil {
			errorMessages = append(errorMessages, matches[1])
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Scanner error: %v", err)
		return nil, err
	}

	return errorMessages, nil
}

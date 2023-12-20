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

type ListOfLogs struct {
	Logs []Log `json:"messages"`
}

type Log struct {
	When    string `json:"when"`
	Message string `json:"message"`
}

func Test_GetLog_ShouldReturnRecentLogEntries_InJsonFormat(t *testing.T) {
	originalApiLog := GetLog(t)
	var apiLogs []Log
	for _, log := range originalApiLog.Logs {
		if log.Message != "..." {
			apiLogs = append(apiLogs, log)
		}
	}
	logFileEntries, err := ParseLogFile("../api-test-home/syncthing.log")
	if err != nil {
		t.Fatalf("Could not parse local log file %v", err)
	}

	apiLogLength := len(apiLogs)
	logFileEntryLength := len(logFileEntries)

	lastFileLogEntries := logFileEntries
	if logFileEntryLength > apiLogLength {
		lastFileLogEntries = logFileEntries[logFileEntryLength-apiLogLength:]
	}

	for i := 0; i < apiLogLength; i++ {
		if !reflect.DeepEqual(lastFileLogEntries[i], apiLogs[i].Message) {
			t.Fatalf("Expected: %v, Got: %v", lastFileLogEntries[i], apiLogs[i].Message)
		}
	}

}

func GetLog(t *testing.T) ListOfLogs {
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
	logURL := "http://" + address + "/rest/system/log"
	response, err := test_api.MakeHttpRequest("GET", apikey, logURL)
	if err != nil {
		t.Fatalf("Could not do post request: %v", err)
	}
	var listOfLogs ListOfLogs
	if err := json.NewDecoder(response.Body).Decode(&listOfLogs); err != nil {
		t.Fatalf("Could not decode response: %v", err)
	}
	return listOfLogs
}

func ParseLogFile(pathToLogFile string) ([]string, error) {
	var re = regexp.MustCompile(`\[(.*?)\] \d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2} (INFO|WARNING|ERROR|.*?): (.*)`)
	file, err := os.Open(pathToLogFile)
	if err != nil {
		log.Fatalf("Could not open log file: %v", err)
		return nil, err
	}
	defer file.Close()

	logEntries := []string{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		matches := re.FindStringSubmatch(scanner.Text())
		if matches != nil {
			logEntries = append(logEntries, matches[3]) // The message is in the third capture group
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Scanner error: %v", err)
		return nil, err
	}

	return logEntries, nil
}

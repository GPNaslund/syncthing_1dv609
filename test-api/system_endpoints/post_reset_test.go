package system_endpoints

/*
package system_endpoints

import (
	test_api "github.com/syncthing/syncthing/test-api"
	"github.com/syndtr/goleveldb/leveldb"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func Test_DB(t *testing.T) {

	dir, err := filepath.Abs("../api-test-home/index-v0.14.0.db")
	if err != nil {
		t.Fatalf("Could not open db")
	}

	db, err := leveldb.OpenFile(dir, nil)
	if err != nil {
		t.Fatalf("Could not open database directory: %v", err)
	}

	err = db.Put([]byte("test"), []byte("test-value"), nil)
	if err != nil {
		t.Fatalf("Could not add test data file")
	}

	_, err = db.Get([]byte("test"), nil)
	if err != nil {
		t.Fatalf("Could not get the new file: %v", err)
	}
	db.Close()
	MakePostRequestToReset(t)
	var newDB *leveldb.DB
	for attempts := 0; attempts < 10; attempts++ {
		newDB, err = leveldb.OpenFile(dir, nil)
		if err == nil {
			break
		}
		time.Sleep(time.Second) // Wait before retrying
	}
	if err != nil {
		t.Fatalf("Could not open database after reset: %v", err)
	}

	_, err = newDB.Get([]byte("test"), nil)
	if err == nil {
		t.Fatal("The testfile is still in db!")
	}

}

func MakePostRequestToReset(t *testing.T) {
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
	resetURL := "http://" + address + "/rest/system/reset"
	_, err = test_api.MakeHttpRequest("POST", apikey, resetURL)
	if err != nil {
		t.Fatalf("Could not do post request: %v", err)
	}

	time.Sleep(time.Second * 10)

	shutdownURL := "http://" + address + "/rest/system/shutdown"
	_, err = test_api.MakeHttpRequest("POST", apikey, shutdownURL)

	if err != nil {
		t.Fatalf("Could not shut down syncthing via api: %v", err)
	}

}
*/

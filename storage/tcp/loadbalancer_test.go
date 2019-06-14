package tcp

import (
	"io"
	"os"
	"testing"
)

func TestWriteQuery(t *testing.T) {
	expected := "0 arg1 arg2 34\n"
	output := craftQuery(SubscribeNew, "arg1", "arg2", "34")
	if output != expected {
		t.Errorf("expected %s, got %s", expected, output)
	}
}

func TestGetId(t *testing.T) {
	file := createTestFileId(t)
	if file == nil {
		t.Errorf("There was a problem setting up the file id test")
		return
	}
	defer cleanUpBackup()

	expected := "testId"
	id := GetId()
	if expected != id {
		t.Errorf("expected %s, got %s", expected, id)
	}

	file.Close()

	err := os.RemoveAll(os.Getenv("STORAGE_ID_FILE"))
	if err != nil {
		t.Errorf("could not delete storage file : %s", err)
	}

	expected = ""
	id = GetId()
	if expected != id {
		t.Errorf("id should be empty")
	}
}

func createTestFileId(t *testing.T) *os.File {
	path := os.Getenv("STORAGE_ID_FILE")
	var file *os.File
	if _, err := os.Stat(path); os.IsNotExist(err) {
		file, err = os.Create(path)
		if err != nil {
			t.Errorf("could not create file in %s : %s", path, err)
			return nil
		}
	} else {
		//Copy id file
		copyIdFile(path)
		file, err = os.OpenFile(path, os.O_WRONLY, 0644)
		if err != nil {
			t.Errorf("could not open file")
		}
	}

	if file != nil {
		_ ,err := file.Write([]byte("testId"))
		if err != nil {
			t.Errorf("could not write testId to test id file : %s", err)
		}
	}
	return file
}

func copyIdFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		panic("could not open id file")
	}
	defer file.Close()
	backup, err := os.Create(path + ".bck")
	if err != nil {
		panic("could not create backup file")
	}
	defer backup.Close()

	_, err = io.Copy(backup, file)
	if err != nil {
		panic("could not copy id file")
	}
}


func cleanUpBackup() {
	path := os.Getenv("STORAGE_ID_FILE")
	os.RemoveAll(path)

	backupPath := path + ".bck"
	//If there is a backup file
	if _, err := os.Stat(backupPath); !os.IsNotExist(err) {

		file, err := os.Create(path)
		if err != nil {
			panic("could not create file")
		}
		defer file.Close()

		bck, err := os.Open(backupPath)
		if err != nil {
			panic("could not open backup file")
		}

		io.Copy(file, bck)
	}

	os.RemoveAll(backupPath)
}

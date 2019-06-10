package main

import (
	"bufio"
	"io"
	"os"
	"strings"
	"testing"
)

func TestSanitarizeString(t *testing.T) {
	healthyString := "ThisIsAHealthyLittleString"
	sanitarized := sanitarizeString(healthyString)

	if healthyString != sanitarized {
		t.Errorf("Nothing should have been removed, " +
			"healty is \"%s\", sanitarized was \"%s\"\n", healthyString, sanitarized)
	}

	stringWithSpace := "This Is A String With Space"
	expected := "ThisIsAStringWithSpace"
	sanitarized = sanitarizeString(stringWithSpace)

	if sanitarized != expected {
		t.Errorf("didn't remove spaces, expected \"%s\", got \"%s\"", expected, sanitarized)
	}

	stringWithDots := "This.Is.A.String.With.Dots"
	expected = "ThisIsAStringWithDots"
	sanitarized = sanitarizeString(stringWithDots)

	if sanitarized != expected {
		t.Errorf("didn't remove dots, expected \"%s\", got \"%s\"", expected, sanitarized)
	}

	stringWithSlashes := "This/Is/A/String/With/Slash"
	expected  = "ThisIsAStringWithSlash"
	sanitarized = sanitarizeString(stringWithSlashes)

	if sanitarized != expected {
		t.Errorf("didn't remove slashes, expected \"%s\", got \"%s\"", expected, sanitarized)
	}

	stringWithAll := "This Is A ..String/With./Lots Of Bad ...././../Stuff"
	expected = "ThisIsAStringWithLotsOfBadStuff"
	sanitarized = sanitarizeString(stringWithAll)

	if sanitarized != expected {
		t.Errorf("didnt sanitarize at all, expected \"%s\", got \"%s\"", expected, sanitarized)
	}
}

func TestParseHash(t *testing.T) {
	validHash := "ca5ab459530e0e928155af72c8fafd74902d7469eff2224a6b3722b000ff6cdb"
	reader := strings.NewReader(validHash)
	parsed, err := parseHash(reader)
	if err != nil {
		t.Errorf("hash should be valid : %s", err.Error())
	}

	if parsed != validHash {
		t.Error("error reading hash")
	}

	tooShortHash := "ca5ab459530e0e928155af72c8fafd74902d7469eff2224a6b3722b000"
	reader = strings.NewReader(tooShortHash)

	parsed, err = parseHash(reader)
	if _, ok := err.(HashTooShort); !ok  || parsed != ""{
		t.Errorf("hash is too short, it should not be valid : %s", err)
	}

	invalidHash := "ca.ab459/30e0e9281 5af72..faf 74902d7469eff2224a6b3722b000ff6cdb"
	reader = strings.NewReader(invalidHash)

	parsed, err = parseHash(reader)
	if _, ok := err.(HashInvalid); !ok || parsed != "" {
		t.Errorf("hash shoud be invalid : %s", err)
	}
}

func TestGetAbsDirectory(t *testing.T) {
	expected := os.Getenv("STORAGE_DIR")
	path := getAbsDirectory()
	if path != expected {
		t.Errorf("The path should be %s, was %s", expected, path)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("The folder should have been created")
	}
}

func TestWriteFileToDisk(t *testing.T) {
	filename := "test"
	fileContent := "this is a file containing all of this text yeaaaah!"
	reader := strings.NewReader(fileContent)

	err := writeFileToDisk(filename, reader, 1000)
	if err != nil {
		t.Errorf("error writing file : %s",  err)
	}

	file, err := os.Open(getAbsDirectory() + filename)
	if err != nil {
		t.Errorf("could not open file : %s", err)
	}
	defer file.Close()

	fileReader := bufio.NewReader(file)
	buf := make([]byte, len([]byte(fileContent)))
	n ,err := fileReader.Read(buf)

	if err != nil && err != io.EOF {
		t.Errorf("could not read file : %s", err)
	}

	if n != len(buf) {
		t.Errorf("didn't read enough lines, expected %d, got %d", len(buf), n)
	}

	content := string(buf[:n])
	if content != fileContent {
		t.Errorf("file read correctly but content was different, expected %s, got %s", fileContent, content)
	}
}
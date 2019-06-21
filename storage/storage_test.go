package main

import (
	"bufio"
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"testing"
	"storage/utils"
)

func TestSanitarizeString(t *testing.T) {
	healthyString := "ThisIsAHealthyLittleString"
	testSubSanitarize(t, healthyString, healthyString)
	testSubSanitarize(t, "This Is A String With Space", "ThisIsAStringWithSpace")
	testSubSanitarize(t, "This.Is.A.String.With.Dots", "ThisIsAStringWithDots")
	testSubSanitarize(t, "This/Is/A/String/With/Slash", "ThisIsAStringWithSlash")
	testSubSanitarize(t, "This Is A ..String/With./Lots Of Bad ...././../Stuff", "ThisIsAStringWithLotsOfBadStuff")
}

func testSubSanitarize(t *testing.T, input string, expected string) {
	sanitarized := sanitarizeString(input)

	if expected != sanitarized {
		t.Errorf("sanitarization failed : expected %s, got %s", expected, sanitarized)
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
	if _, ok := err.(HashTooShort); !ok || parsed != "" {
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
	expected := utils.GetRequiredEnv("STORAGE_DIR")
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
		t.Errorf("error writing file : %s", err)
	}

	path := getAbsDirectory() + filename
	file, err := os.Open(path)
	if err != nil {
		t.Errorf("could not open file : %s", err)
	}

	//Deleting file when done
	defer func() {
		err := os.RemoveAll(path)
		if err != nil {
			t.Errorf("could not delete file !")
		}
	}()
	defer file.Close()

	testFileContent(t, []byte(fileContent), file)
}

func testFileContent(t *testing.T, expected []byte, reader io.Reader) {
	fileReader := bufio.NewReader(reader)
	buf := make([]byte, len(expected))
	n, err := fileReader.Read(buf)

	if err != nil && err != io.EOF {
		t.Errorf("could not read file : %s", err)
		return
	}

	if n != len(buf) {
		t.Errorf("didn't read enough lines, expected %d, got %d", len(buf), n)
		return
	}

	content := buf[:n]

	for i, val := range content {
		if buf[i] != val {
			t.Errorf("file read correctly but byte %d was different, expected %b, got %b", i, buf[i], val)
		}
	}
}

func TestParseMultiPart(t *testing.T) {
	body := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(body)
	hash := "8b656fc7d92aa7a479f1ff34e22ac1e75c56dc96fc376c0fa2546cca1e3f2409"
	err := bodyWriter.WriteField("hash", hash)
	if err != nil {
		t.Errorf("could not write filed \"hash\" to body : %s", err)
	}

	fileWriter, err := bodyWriter.CreateFormFile("file", "random")
	if err != nil {
		t.Errorf("could not create form file : %s", err)
	}

	fileContent := []byte("test")

	fileWriter.Write(fileContent)
	bodyWriter.Close()

	req, err := http.NewRequest(http.MethodGet, "", body)
	req.Header.Set("Content-Type", bodyWriter.FormDataContentType())

	if err != nil {
		t.Errorf("could not create request : %s", err)
		return
	}

	req.ParseMultipartForm(3500)

	filename, _, err := parseMultiPartForm(req)
	if err != nil {
		t.Errorf("could not parse multipartform : %s", err)
	}

	if filename != hash {
		t.Errorf("hash should have been parse : expected %s, got %s", hash, filename)
	}

	file, _, err := req.FormFile("file")
	if err != nil {
		t.Errorf("could not read file from req")
	}
	testFileContent(t, fileContent, file)
}

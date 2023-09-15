package main

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

const fileStoragePathTest = "./testDir"

var app = NewApp(fileStoragePathTest)

func TestSaveFileHandler(t *testing.T) {
	if err := os.MkdirAll(fileStoragePathTest, 0755); err != nil {
		t.Fatalf("Failed to create the directory: %v", err)
	}

	fileContents := "Hello, world!"
	fileName := "testfile.txt"
	filePath := "/save/" + fileName

	buffer := &bytes.Buffer{}
	writer := multipart.NewWriter(buffer)
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	_, err = io.Copy(part, bytes.NewReader([]byte(fileContents)))
	if err != nil {
		t.Fatalf("Failed to write file contents: %v", err)
	}

	err = writer.Close()
	if err != nil {
		t.Fatalf("Failed to close: %v", err)
	}
	request, err := http.NewRequest("POST", filePath, buffer)
	if err != nil {
		t.Fatalf("Failed to create POST request: %v", err)
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())

	recorder := httptest.NewRecorder()

	app.saveFileHandler(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, but got %d", http.StatusCreated, recorder.Code)
	}

	expectedResponse := "File saved successfully"
	if recorder.Body.String() != expectedResponse {
		t.Errorf("Expected response body %q, but got %q", expectedResponse, recorder.Body.String())
	}

	fullPath := filepath.Join(fileStoragePathTest, fileName)
	contents, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	if string(contents) != fileContents {
		t.Errorf("Expected file contents %q, but got %q", fileContents, string(contents))
	}
}

func TestServeFileHandler(t *testing.T) {
	fileContents := "Hello, world!"
	fileName := "testfile.txt"
	filePath := filepath.Join(fileStoragePathTest, fileName)
	err := os.WriteFile(filePath, []byte(fileContents), 0644)
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}

	req, err := http.NewRequest("GET", "/serve/"+fileName, nil)
	if err != nil {
		t.Fatalf("Failed to create GET request: %v", err)
	}

	recorder := httptest.NewRecorder()

	app.serveFileHandler(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, recorder.Code)
	}

	expectedResponse := fileContents
	if recorder.Body.String() != expectedResponse {
		t.Errorf("Expected response body %q, but got %q", expectedResponse, recorder.Body.String())
	}
}

func TestDeleteFileHandler(t *testing.T) {
	fileContents := "Hello, world!"
	fileName := "testfile.txt"
	filePath := filepath.Join(fileStoragePathTest, fileName)
	err := os.WriteFile(filePath, []byte(fileContents), 0644)
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}

	request, err := http.NewRequest("DELETE", "/delete/"+fileName, nil)
	if err != nil {
		t.Fatalf("Failed to create DELETE request: %v", err)
	}

	recorder := httptest.NewRecorder()

	app.deleteFileHandler(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected the status code %d, but got %d", http.StatusOK, recorder.Code)
	}

	expectedResponse := "File deleted successfully"
	if recorder.Body.String() != expectedResponse {
		t.Errorf("Expected response body %q, but got %q", expectedResponse, recorder.Body.String())
	}

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Errorf("Expected the file %q to be deleted, but it still exists", filePath)
	}
}

func TestWelcomeHandler(t *testing.T) {
	request, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Failed to create GET request: %v", err)
	}

	recorder := httptest.NewRecorder()
	app.welcomeHandler(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected the status code %d, but got %d", http.StatusOK, recorder.Code)
	}

	expectedResponce := `
    <!DOCTYPE html>
    <html>
    <head>
        <title>Simple File Server</title>
    </head>
    <body>
        <h1>Welcome to the File Server</h1>
        <p>you can do these things below:</p>
        <ul>
            <li>Upload files via POST request to /save/</li>
            <li>Retrieve files via GET request to /serve/</li>
            <li>Delete files via DELETE request to /delete/</li>
        </ul>
    </body>
    </html>
    `
	if recorder.Body.String() != expectedResponce {
		t.Errorf("can't match HTML response")
	}
}

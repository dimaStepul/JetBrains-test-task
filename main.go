package main

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

var directoryPath = "./files"

const filePermission = 0755

type App struct {
	fileStoragePath string
}

func NewApp(fileStoragePath string) *App {
	return &App{
		fileStoragePath: fileStoragePath,
	}
}

func (app *App) saveFileHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(responseWriter, "Invalid HTTP method, use POST", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := request.FormFile("file")
	if err != nil {
		http.Error(responseWriter, fmt.Sprintf("Failed to read the uploaded file: %v", err), http.StatusInternalServerError)
		return
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			log.Printf("Failed to close the file: %v", err)
		}
	}(file)

	filePath := request.URL.Path[len("/save/"):]

	fullPath := filepath.Join(app.fileStoragePath, filePath)

	directory := filepath.Dir(fullPath)
	if err := os.MkdirAll(directory, 0755); err != nil {
		http.Error(responseWriter, fmt.Sprintf("Failed to create directories: %v", err), http.StatusInternalServerError)
		return
	}

	destination, err := os.Create(fullPath)
	if err != nil {
		http.Error(responseWriter, fmt.Sprintf("Failed to create the file: %v", err), http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := destination.Close(); err != nil {
			log.Printf("Failed to close the file: %v", err)
		}
	}()
	if _, err := io.Copy(destination, file); err != nil {
		http.Error(responseWriter, fmt.Sprintf("Failed to save the file: %v", err), http.StatusInternalServerError)
		return
	}

	responseWriter.WriteHeader(http.StatusCreated)
	_, err = responseWriter.Write([]byte("File saved successfully"))
	if err != nil {
		http.Error(responseWriter, "Failed to response", http.StatusInternalServerError)
	}
}

func (app *App) serveFileHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		http.Error(responseWriter, "Invalid HTTP method, expected GET", http.StatusMethodNotAllowed)
		return
	}

	filePath := request.URL.Path[len("/serve/"):]

	fullPath := filepath.Join(app.fileStoragePath, filePath)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		http.Error(responseWriter, "File is  not found", http.StatusNotFound)
		return
	}

	http.ServeFile(responseWriter, request, fullPath)
}

func (app *App) deleteFileHandler(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodDelete {
		http.Error(responseWriter, "Invalid HTTP method, expected DELETE", http.StatusMethodNotAllowed)
		return
	}

	filePath := request.URL.Path[len("/delete/"):]

	fullPath := filepath.Join(app.fileStoragePath, filePath)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		http.Error(responseWriter, "File is not found", http.StatusNotFound)
		return
	}

	if err := os.Remove(fullPath); err != nil {
		http.Error(responseWriter, fmt.Sprintf("Failed to delete the file: %v", err), http.StatusInternalServerError)
		return
	}

	_, err := responseWriter.Write([]byte("File deleted successfully"))
	if err != nil {
		http.Error(responseWriter, "Failed to response", http.StatusInternalServerError)
	}
}

func (app *App) welcomeHandler(responseWriter http.ResponseWriter, request *http.Request) {
	html := `
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
	responseWriter.WriteHeader(http.StatusOK)
	responseWriter.Header().Set("Content-Type", "text/html")
	_, err := responseWriter.Write([]byte(html))
	if err != nil {
		http.Error(responseWriter, "Failed to response", http.StatusInternalServerError)
	}
}
func main() {
	if err := os.MkdirAll(directoryPath, filePermission); err != nil {
		log.Fatalf("Failed to create the directory: %v", err)
	}

	app := NewApp(directoryPath)

	http.HandleFunc("/", app.welcomeHandler)
	http.HandleFunc("/save/", app.saveFileHandler)
	http.HandleFunc("/serve/", app.serveFileHandler)
	http.HandleFunc("/delete/", app.deleteFileHandler)
	
	httpPort, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		fmt.Println("Can't convert from string to int ")
	}
	if httpPort == 0 {
		httpPort = 9999
	}
	
	fmt.Println("File server is running")

	err = http.ListenAndServe(fmt.Sprintf(":%d", httpPort), nil)

	if err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

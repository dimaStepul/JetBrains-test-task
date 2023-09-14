package main

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

const fileStoragePath = "./files"

func saveFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid HTTP method, expected POST", http.StatusMethodNotAllowed)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read the uploaded file: %v", err), http.StatusInternalServerError)
		return
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			log.Printf("Failed to close the file: %v", err)
		}
	}(file)

	filePath := r.URL.Path[len("/save/"):]

	fullPath := filepath.Join(fileStoragePath, filePath)

	directory := filepath.Dir(fullPath)
	if err := os.MkdirAll(directory, 0755); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create directories: %v", err), http.StatusInternalServerError)
		return
	}

	destination, err := os.Create(fullPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create the file: %v", err), http.StatusInternalServerError)
		return
	}
	defer func() {
		if err := destination.Close(); err != nil {
			log.Printf("Failed to close the file: %v", err)
		}
	}()
	if _, err := io.Copy(destination, file); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save the file: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte("File saved successfully"))
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

func serveFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid HTTP method, expected GET", http.StatusMethodNotAllowed)
		return
	}

	filePath := r.URL.Path[len("/serve/"):]

	fullPath := filepath.Join(fileStoragePath, filePath)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, fullPath)
}

func deleteFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid HTTP method, expected DELETE", http.StatusMethodNotAllowed)
		return
	}

	filePath := r.URL.Path[len("/delete/"):]

	fullPath := filepath.Join(fileStoragePath, filePath)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	if err := os.Remove(fullPath); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete the file: %v", err), http.StatusInternalServerError)
		return
	}

	_, err := w.Write([]byte("File deleted successfully"))
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

func welcomeHandler(w http.ResponseWriter, r *http.Request) {
	html := `
    <!DOCTYPE html>
    <html>
    <head>
        <title>File Server</title>
    </head>
    <body>
        <h1>Welcome to the File Server</h1>
        <p>This file server allows you to:</p>
        <ul>
            <li>Upload files via POST request to /save/</li>
            <li>Retrieve files via GET request to /serve/</li>
            <li>Delete files via DELETE request to /delete/</li>
        </ul>
    </body>
    </html>
    `
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")
	_, err := w.Write([]byte(html))
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

func main() {
	if err := os.MkdirAll(fileStoragePath, 0755); err != nil {
		log.Fatalf("Failed to create file storage directory: %v", err)
	}

	http.HandleFunc("/", welcomeHandler)
	http.HandleFunc("/save/", saveFileHandler)
	http.HandleFunc("/serve/", serveFileHandler)
	http.HandleFunc("/delete/", deleteFileHandler)

	fmt.Println("File server is running on localhost:9999")
	err := http.ListenAndServe("localhost:9999", nil)
	if err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"time"

	wkhtml "github.com/SebastiaanKlippert/go-wkhtmltopdf"
)

const MAX_UPLOAD_SIZE = 1024 * 1024 * 2000 //2GB
const HTML_FILE_NAME = "index.html"
const UPLOADS_DIR = "./uploads"
const UNZIP_DIR = "./unziped"
const LOG_FILE = "./logs"
const PDF_DIR = "./pdfs"

func uploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		log.Println("Method not allowed")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, MAX_UPLOAD_SIZE)
	if err := r.ParseMultipartForm(MAX_UPLOAD_SIZE); err != nil {
		http.Error(w, "The uploaded file is too big. Please choose a file that's less than 2GB in size", http.StatusBadRequest)
		log.Println("The uploaded file is too big.")
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Println(err)
		return
	}
	defer file.Close()

	// create buffer to determine the MIME type
	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	filetype := http.DetectContentType(buff)
	if filetype != "application/zip" {
		http.Error(w, "The provided file format is not allowed. Please upload a zip archive", http.StatusBadRequest)
		log.Println("The provided file format is not allowed.")
		return
	}

	// return pointer back to the start of the file
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	dst, err := os.Create(path.Join(UPLOADS_DIR, handler.Filename))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	defer dst.Close()
	
	// Copy the uploaded file to the created file on the filesystem
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	log.Printf("Uploaded file: %+v\n", handler.Filename)
	log.Printf("File size: %+v\n", handler.Size)

	fmt.Fprintln(w, "File was uploaded successfully.")

	unzip(dst.Name(), UNZIP_DIR)
	convertHtml2Pdf(UNZIP_DIR, PDF_DIR)
}

func unzip(source, dest string) error {
	read, err := zip.OpenReader(source)
	if err != nil {
		log.Println(err)
		return err
	}
	defer read.Close()

	for _, file := range read.File {
		if file.Mode().IsDir() {
			continue
		}
		open, err := file.Open()
		if err != nil {
			log.Println(err)
			return err
		}
		name := path.Join(dest, file.Name)
		create, err := os.Create(name)
		if err != nil {
			log.Println(err)
			return err
		}
		defer create.Close()
		create.ReadFrom(open)
	}
	return nil
}

func convertHtml2Pdf(source, dest string) {
	start := time.Now()
	heapStart := getHeapSize()

	pdfg, err := wkhtml.NewPDFGenerator()
	if err != nil {
		log.Println(err)
		return
	}

	file, err := os.Open(path.Join(source, HTML_FILE_NAME))
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	pdfg.AddPage(wkhtml.NewPageReader(file))

	// Create PDF document in internal buffer
	err = pdfg.Create()
	if err != nil {
		log.Println(err)
		return
	}

	// date and time is in YYYYMMDDhhmmss format
	pdfFileName := time.Now().Format("20060102150405") + ".pdf"

	// Write buffer contents to file on disk
	err = pdfg.WriteFile(path.Join(dest, pdfFileName))
	if err != nil {
		log.Println(err)
		return
	}

	elapsed := time.Since(start)
	heapEnd := getHeapSize()
	heapUsed := heapEnd - heapStart

	log.Printf("Converted file: %s. Time taken: %s. Memory used: %v bytes", pdfFileName, elapsed, heapUsed)
}

func getHeapSize() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc
}

func main() {
	logFile, err := os.OpenFile(LOG_FILE, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)

	if err != nil {
		log.Println(err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	err = os.MkdirAll(UPLOADS_DIR, os.ModePerm)
	if err != nil {
		log.Println(err)
		return
	}

	err = os.MkdirAll(UNZIP_DIR, os.ModePerm)
	if err != nil {
		log.Println(err)
		return
	}

	err = os.MkdirAll(PDF_DIR, os.ModePerm)
	if err != nil {
		log.Println(err)
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/upload", uploadFile)

	err = http.ListenAndServe("0.0.0.0:3333", mux)

	if err != nil {
		log.Printf("Error starting server: %s\n", err)
		os.Exit(1)
	}
}

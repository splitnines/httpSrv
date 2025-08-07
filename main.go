package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	HttpSrv()
}

func HttpSrv() {
	http.HandleFunc("/", logRequest(fileHandler))

	ip := getLocalIP()
	port := ":12345"
	fmt.Printf("[+] Running Extended HTTP Server on %s port %s\n", ip, port[1:])
	fmt.Printf("[+] Server URL: (http://%s%s/)\n", ip, port)
	fmt.Println("[+] Press Ctrl-c to stop the server")
	log.Fatal(http.ListenAndServe(port, nil))
}

func fileHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		Get(w, r)
	case http.MethodPut:
		Put(w, r)
	default:
		http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
	}
}

func Get(w http.ResponseWriter, r *http.Request) {
	path := filepath.Clean("." + r.URL.Path)
	http.ServeFile(w, r, path)
}

func Put(w http.ResponseWriter, r *http.Request) {
	path := filepath.Clean("." + r.URL.Path)

	if strings.HasSuffix(r.URL.Path, "/") {
		http.Error(w, "Don't even try to put that directory", http.StatusBadRequest)
		return
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		http.Error(w, "Failed to create directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	file, err := os.Create(path)
	if err != nil {
		http.Error(w, "Shit, failed to create the file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	if _, err := io.Copy(file, r.Body); err != nil {
		http.Error(w, "Shit, failed to write file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "localhost"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func logRequest(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)

		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(lrw, r)

		log.Printf("[%s] %s %s - %d (%s)", ip, r.Method, r.URL.Path, lrw.statusCode, time.Since(start))
	}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHandler(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

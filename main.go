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
	fmt.Printf("[+] Running Extended HTTP Server on %s port %s\n",
		ip, port[1:])
	fmt.Printf("[+] Server URL: (http://%s%s/)\n", ip, port)
	fmt.Println("[+] Press Ctrl-c to stop the server")
	go spinner()
	log.Fatal(http.ListenAndServe(port, nil))
}

func fileHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		Get(w, r)
	case http.MethodPut:
		Put(w, r)
	default:
		http.Error(w, "Unsupported method",
			http.StatusMethodNotAllowed)
	}
}

func Get(w http.ResponseWriter, r *http.Request) {
	path := filepath.Clean("." + r.URL.Path)
	if strings.Contains(path, "/") {
		http.Error(w, "These are not the driods you're looking for.",
			http.StatusForbidden)
		return
	}
	if ok := strings.Compare(path, "."); ok == 0 {
		http.Error(w, "Don't be so nosey.",
			http.StatusForbidden)
		return
	}

	http.ServeFile(w, r, path)
}

func Put(w http.ResponseWriter, r *http.Request) {
	path := filepath.Clean("." + r.URL.Path)

	if strings.Contains(path, "/") {
		http.Error(w, "Don't even try to put to a subdirectory",
			http.StatusBadRequest)
		return
	}

	file, err := os.Create(path)
	if err != nil {
		http.Error(w, "Shit, failed to create the file: "+err.Error(),
			http.StatusInternalServerError)
		return
	}
	defer file.Close()

	if _, err := io.Copy(file, r.Body); err != nil {
		http.Error(w, "Shit, failed to write file: "+err.Error(),
			http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:53")
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

		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		next.ServeHTTP(lrw, r)

		fmt.Print("\b")
		log.Printf("[%s] %s %s - %d (%s)", ip, r.Method, r.URL.Path,
			lrw.statusCode, time.Since(start))
	}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func spinner() {
	spinChars := []rune{'|', '/', '-', '\\'}
	i := 0
	for {
		fmt.Printf("\r%c", spinChars[i%len(spinChars)])
		i = (i + 1) % len(spinChars)
		time.Sleep(100 * time.Millisecond)
	}
}

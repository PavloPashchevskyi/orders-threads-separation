package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// spaHandler implements the http.Handler interface, so we can use it
// to respond to HTTP requests. The path to the static directory and
// path to the index file within that static directory are used to
// serve the SPA in the given static directory.
type spaHandler struct {
	staticPath string
	indexPath  string
}

// ServeHTTP inspects the URL path to locate a file within the static dir
// on the SPA handler. If a file is found, it will be served. If not, the
// file located at the index path on the SPA handler will be served. This
// is suitable behavior for serving an SPA (single page application).
func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get the absolute path to prevent directory traversal
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		// if we failed to get the absolute path respond with a 400 bad request
		// and stop
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// prepend the path with the path to the static directory
	path = filepath.Join(h.staticPath, path)

	// check whether a file exists at the given path
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		// file does not exist, serve index.html
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	} else if err != nil {
		// if we got an error (that wasn't that the file doesn't exist) stating the
		// file, return a 500 internal server error and stop
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// otherwise, use http.FileServer to serve the static dir
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}

func request(method string, address interface{}, reader io.Reader) map[string]string {
	request, _ := http.NewRequest(method, address.(string), reader)
	client := &http.Client{}
	resp, _ := client.Do(request)

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	defer resp.Body.Close()
	dataJson := buf.String()
	var str map[string]string
	fmt.Println(dataJson)
	fmt.Println("=========================================")
	json.Unmarshal([]byte(dataJson), &str)
	return str
}

func getJSONItem(data map[string]interface{}, key int, wg * sync.WaitGroup) {
	threads := data["threads"]
	thread := threads.([]interface{})[key]

	for i := 0; i < len(thread.([]interface{})); i++ {
		threadToProcess := thread.([]interface{})[i]
		// get request data
		requestData := threadToProcess.(map[string]interface{})["account"]
		requestString := fmt.Sprintf("%v", requestData)
		resp := request("PUT", data["address"], strings.NewReader(requestString))
		// send order request
		if status, ok := resp["status"]; ok && status == "O" {
			orderRequestData := threadToProcess.(map[string]interface{})["order"]
			orderRequestString := fmt.Sprintf("%v", orderRequestData)
			request("PUT", data["address"], strings.NewReader(orderRequestString))
		}
	}
	wg.Done()
}

func routine(data map[string]interface{}) {
	var wg sync.WaitGroup
	threadsCount := len(data["threads"].([]interface{}))

	for i := 0; i < threadsCount; i++ {
		wg.Add(1)
		go getJSONItem(data, i, &wg)
	}

	func() {
		wg.Wait()
		fmt.Println("All threads have been ended!")
		requestString := fmt.Sprintf("%v", data["response"])
		request("PUT", data["address"], strings.NewReader(requestString))
	}()
}

func main() {

	router := mux.NewRouter()

	router.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	router.HandleFunc("/api/multiorder", func(w http.ResponseWriter, r *http.Request) {

		var multiOrderRequestMap map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&multiOrderRequestMap)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		routine(multiOrderRequestMap)
	}).Methods("POST")

	router.HandleFunc("/api/items", func(w http.ResponseWriter, r *http.Request) {
		//data := request()
		//fmt.Println(data)
		//routine(data)
		//// an example API handler
		//w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		//json.NewEncoder(w).Encode(data)
	}).Methods("POST", "PUT")

	spa := spaHandler{staticPath: "build", indexPath: "index.html"}
	router.PathPrefix("/").Handler(spa)

	srv := &http.Server{
		Handler: router,
		Addr:    "127.0.0.1:8000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

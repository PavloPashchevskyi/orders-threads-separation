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
	json.Unmarshal([]byte(dataJson), &str)
	return str
}

func getJSONItem(data map[string]interface{}, key int)  {
	threads := data["threads"]
	thread := threads.([]interface{})[key]

	threadToProcess := thread.([]interface{})[key]

	requestData := threadToProcess.(map[string]interface{})["account"]
	requestString := fmt.Sprintf("%v", requestData)
	ioReader := strings.NewReader(requestString)
	respMap := request("PUT", data["address"], ioReader)
	fmt.Println(respMap)
	//
	//return respMap
}

func routine(data map[string]interface{}) {
	for i := 0; i < len(data); i++ {
		go getJSONItem(data, i)
	}
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
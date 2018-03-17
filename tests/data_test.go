package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
        "../config/mappings"
	"gopkg.in/yaml.v2"
)

const (
    testyaml = "trek/trek_drilldata.yml"
)


func post(testdata []string) {
    file := ioutil.ReadFile(testdata)

    handler := func(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, file)
    }

    req := httptest.NewRequest("GET", "http://localhost:8000/data", nil)
    w := httptest.NewRecorder()
    handler(w, req)

    resp := w.Result()
    body, _ := ioutil.ReadAll(resp.Body)

    fmt.Println(resp.StatusCode)
    fmt.Println(resp.Header.Get("Content-Type"))
    fmt.Println(string(body))
}


func Sum(x int, y int) int {
    return x + y
}


func main() {
    post(testyaml)
}

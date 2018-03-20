package main

import (
	"bytes"
        "fmt"
	"io"
        "log"
        "mime/multipart"
	"net/http"
	//"net/http/httptest"
        "os"
        "path/filepath"
)

const (
    testyaml = "tests/trek/trek_drilldata.yml"
    testdata = "tests/trek/trek_drilldata.csv"
    BaseUrl = "http://localhost:8000/data"
)


func newfileConversionRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
    file, err := os.Open(path)
    if err != nil {
      return nil, err
    }
    defer file.Close()

    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    part, err := writer.CreateFormFile(paramName, filepath.Base(path))
    if err != nil {
      return nil, err
    }
    _, err = io.Copy(part, file)

    for key, val := range params {
      _ = writer.WriteField(key, val)
    }
    err = writer.Close()
    if err != nil {
      return nil, err
    }

    req, err := http.NewRequest("POST", uri, body)
    req.Header.Set("Content-Type", writer.FormDataContentType())
    return req, err
}


func main() {
    extraParams := map[string]string{
      "info":  testyaml,
    }
    request, err := newfileConversionRequest(BaseUrl, extraParams, "file", testdata)
    if err != nil {
      log.Fatal(err)
    }
    client := &http.Client{}
    resp, err := client.Do(request)
    if err != nil {
      log.Fatal(err)
    } else {
      body := &bytes.Buffer{}
      _, err := body.ReadFrom(resp.Body)
      if err != nil {
          log.Fatal(err)
      }
      resp.Body.Close()
        fmt.Println(resp.StatusCode)
        fmt.Println(resp.Header)
        fmt.Println(body)
    }
}

package main

import (
	"bytes"
        "fmt"
        "io/ioutil"
        "mime/multipart"
	"net/http"
        "os"
        "path/filepath"
)

const (
    testjson = "trek/trek_drilldata.json"
    testdata = "trek/trek_drilldata.csv"
    testurl = "http://mapp.life:8000/data"
)

func check (err error) {
    if err != nil {
      fmt.Printf("Damn, there's an error: %s\n",err)
    }
}


func newfileConversionRequest(uri string, params map[string]string) (*http.Request, error) {

    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)

    for k, v := range params {
      switch k {
        case "info" :
          file, err := os.Open(v)
          check(err)
          stuff, err := ioutil.ReadAll(file)
          check(err)
          part, err := writer.CreateFormFile(k, filepath.Base(v))
          check(err)
          part.Write(stuff)
          fmt.Printf("info written\n")
        case "file" :
          file, err := os.Open(v)
          check(err)
          stuff, err := ioutil.ReadAll(file)
          check(err)
          part, err := writer.CreateFormFile(k, filepath.Base(v))
          check(err)
          part.Write(stuff)
          fmt.Printf("data written\n")
      }
    }

    err := writer.Close()
    if err != nil {
      return nil, err
    }

    req, err := http.NewRequest("POST", uri, body)
    check(err)

    req.Header.Set("Content-Type", writer.FormDataContentType())

    return req, err
}


func main() {

    extraParams := map[string]string{
      "info":  testjson,
      "file":  testdata,
    }

    request, err := newfileConversionRequest(testurl, extraParams)
    if err != nil {
      fmt.Printf("FUNC returns an error: %s\n",err)
    }
    client := &http.Client{}
    fmt.Printf("Request Headers: %s\n",request.Header)
    resp, err := client.Do(request)
    if err != nil {
      fmt.Printf("Request returns an error: %s\n",err)
    } else {
      body := &bytes.Buffer{}
      _, err := body.ReadFrom(resp.Body)
      if err != nil {
          fmt.Printf("Body READ returns an error: %s\n",err)
      }
      resp.Body.Close()
        fmt.Printf("Response Status: %s\n",resp.StatusCode)
        fmt.Printf("Response Headers: %s\n",resp.Header)
        fmt.Printf("Response Body: %s\n",body)
    }
}

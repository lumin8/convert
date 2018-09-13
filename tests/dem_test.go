package main

import (
	"bytes"
	"fmt"
	"net/http"
)

const (
	x       = "-111.02523"
	y       = "45.63856"
	format  = "json"
	testurl = "http://mapp.life:8000/dem"
)

func check(err error) {
	if err != nil {
		fmt.Printf("Damn, there's an error: %s\n", err)
	}
}

func newfileConversionRequest(uri string, params map[string]string) (*http.Request, error) {

	req, err := http.NewRequest("POST", uri, nil)
	//req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	check(err)

	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}

	req.URL.RawQuery = q.Encode()

	return req, err
}

func main() {

	extraParams := map[string]string{
		"x": x,
		"y": y,
		"f": format,
	}

	request, err := newfileConversionRequest(testurl, extraParams)
	if err != nil {
		fmt.Printf("FUNC returns an error: %s\n", err)
	}
	client := &http.Client{}
	fmt.Printf("Request Headers: %s\n", request.Header)
	resp, err := client.Do(request)
	if err != nil {
		fmt.Printf("Request returns an error: %s\n", err)
	} else {
		body := &bytes.Buffer{}
		_, err := body.ReadFrom(resp.Body)
		if err != nil {
			fmt.Printf("Body READ returns an error: %s\n", err)
		}
		resp.Body.Close()
		fmt.Printf("Response Status: %s\n", resp.StatusCode)
		fmt.Printf("Response Headers: %s\n", resp.Header)
		fmt.Printf("Response Body: %s\n", body)
	}
}

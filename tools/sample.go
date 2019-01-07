package main

import (
	"io/ioutil"
	"log"
	"net/http"
)

const (
	collection = "samplecollection"
	tests      = "../tests/"
)

func sampleHandler(w http.ResponseWriter, r *http.Request) {

	get, err := paramCheck("get", r)
	if err != nil {
		w.Write(err)
		r.Body.Close()
		return
	}

	token, err := paramCheck("token", r)
	if err != nil {
		token = "0"
		//w.Write(err)
		//r.Body.Close()
		//return
	}

	did, err := paramCheck("did", r)
	if err != nil {
		did = "0"
		//w.Write(err)
		//r.Body.Close()
		//return
	}

	log.Printf("%v", token)

	switch get {

	case "collection":

		log.Println("sample collection requested")

		f, err := ioutil.ReadFile(tests + collection + ".json")
		if err != nil {
			log.Printf("%s", err)
		}

		w.Write(f)

		r.Body.Close()

	case "dataset":

		if did == "0" {
			w.Write(err)
			r.Body.Close()
			return
		}

		log.Println("sample dataset", did, "requested")

		f, err := ioutil.ReadFile(tests + did + ".json")
		if err != nil {
			log.Printf("%s", err)
		}

		w.Write(f)

		r.Body.Close()

	default:

		w.Write([]byte("please specifiy a 'get' parameter"))
		r.Body.Close()
	}
}

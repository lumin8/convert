package main

import (
    "io/ioutil"
    "log"
    "net/http"
)

const (
    // sample locations
    collection = "tests/samplecollection.json"
    tests = "tests/"
    dataset1 = "tests/sampledataset1.json"
    dataset2 = "tests/sampledataset2.json"
    dataset3 = "tests/sampledataset3.json"
)


func sampleHandler(w http.ResponseWriter, r *http.Request) {
    log.Println("sample project requested")

    get, err := paramCheck("get", r)
    if err != nil {
        w.Write(err)
        r.Body.Close()
        return
    }

    uid, err := paramCheck("uid", r)
    if err != nil {
        w.Write(err)
        r.Body.Close()
        return
    }

    did, err := paramCheck("did", r)
    if err != nil {
        w.Write(err)
        r.Body.Close()
        return
    }

    log.Println("user:",uid)


    switch get {

      case "collection":

        f, err := ioutil.ReadFile(collection)
        if err != nil {
          log.Printf("%s",err)
        }

        w.Write(f)

        r.Body.Close()

      case "dataset":

        f, err := ioutil.ReadFile(tests + did)
        if err != nil {
          log.Printf("%s",err)
        }

        w.Write(f)

        r.Body.Close()

      default:

        w.Write([]byte("incorrect parameters specified"))
        r.Body.Close()
    }
}


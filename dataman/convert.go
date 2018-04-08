package main

import (
    "bytes"
    "encoding/csv"
    "encoding/json"
    "fmt"
    "io"
    "io/ioutil"
    "log"
    "mime"
    "mime/multipart"
    "net/http"
    "strconv"
    "time"
    "gopkg.in/yaml.v2"
)


func dataHandler(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    log.Println("woohoo, got a request!")

    _, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
    check(err)

    reader := multipart.NewReader(r.Body, params["boundary"])

    var indataset Input
    var converted []byte
    var data []byte
    var info []byte

    for {
      part, err := reader.NextPart()
      if err == io.EOF {
        break
      }

      switch part.FormName() {
        case "info" :
          info, err = ioutil.ReadAll(part)
          check(err)
          fmt.Printf("%s\n",info)
          fmt.Printf("read from info!\n")
        case "file" :
          data, err = ioutil.ReadAll(part)
          check(err)
          fmt.Printf("read from data!\n")
      }
    }

    err = yaml.Unmarshal(info, &indataset)
    check(err)

    format := indataset.Format

    switch format {
      case "csv" :
        converted, err = CsvHandler(indataset, data)
        check(err)
      //TBD case "shp": 
        //outdataset, err = ShpHandler(indataset, contents)
      //TBD case "dxf": 
        //outdataset, err = DxfHandler(indataset, contents)
      default :
        converted = []byte("Sorry, things didn't work out.  Is the format supported?")
    }

    ioutil.WriteFile("tests/out.json", converted, 0644)
    w.Write(converted)

    log.Println("total dataset round trip:",int64(time.Since(start).Seconds()),"s")
}


func CsvHandler(indataset Input, contents []byte) (converted []byte, err error) {
    start := time.Now()
    s := bytes.NewReader(contents)

    raw, err := csv.NewReader(s).ReadAll()
    check(err)

    xfield := indataset.Xfield
    yfield := indataset.Yfield
    zfield := indataset.Zfield

    var outdataset Datasets
    var headers map[int]string
    headers = make(map[int]string)

    var attributes Attributes

    for i, record := range raw {
      var point Points
      switch i {
        case 0 :
          for i, header := range record {
            switch header {
              case xfield: headers[i] = "X"
              case yfield: headers[i] = "Y"
              case zfield: headers[i] = "Z"
              default: headers[i] = header
            }
          }
        default :
          for i, value := range record {
            switch headers[i] {
              case "X": point.X, _ = strconv.ParseFloat(value, 64)
              case "Y": point.Y, _ = strconv.ParseFloat(value, 64)
              case "Z": point.Z, _ = strconv.ParseFloat(value, 64)
              default :
                attributes.Key = headers[i]
                attributes.Value = value
                point.Attributes = append(point.Attributes, attributes)
            }
          }
      }

      outdataset.Points = append(outdataset.Points, point)

    }

    converted, err = json.Marshal(outdataset)
    counter.Incr("csv")
    log.Println("csv's processed:",counter.Get("csv"),", time:",int64(time.Since(start).Seconds()),"s")
    return converted, err
}


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
//    "github.com/jonas-p/go-shp"
)


func dataHandler(w http.ResponseWriter, r *http.Request) {
    start := time.Now()

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
      //case "shp":
        //outdataset, err = ShpHandler(indataset, data)
      //case "dxf": 
        //outdataset, err = DxfHandler(indataset, contents)
      default :
        converted = []byte("Sorry, things didn't work out.  Is the format supported?")
    }

    //ioutil.WriteFile("tests/out.json", converted, 0644)
    w.Write(converted)

    log.Println("total dataset round trip:",int64(time.Since(start).Seconds()*1e3),"ms")
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

    for i, record := range raw {
      var pointxyz Point
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
              case "X": pointxyz.X, _ = strconv.ParseFloat(value, 64)
              case "Y": pointxyz.Y, _ = strconv.ParseFloat(value, 64)
              case "Z": pointxyz.Z, _ = strconv.ParseFloat(value, 64)
              default :
                pair := make(map[string]interface{})
                pair[headers[i]] = value
                //attributes.Value = value
                point.Attributes = append(point.Attributes, pair)
            }
          }

          // fill elevation if required
          if pointxyz.Z == 0 && pointxyz.X != 0 && pointxyz.Y != 0 {
            log.Printf("value needed filling in with elevation...")
            pointxyz.Z, err = getElev(pointxyz.X,pointxyz.Y)
          }

          //finally, fill in the point float array
          point.Point = append(point.Point, pointxyz.X, pointxyz.Y, pointxyz.Z)
      }

      outdataset.Points = append(outdataset.Points, point)
    }

    converted, err = json.Marshal(outdataset)
    counter.Incr("csv")
    log.Println("csv's processed:",counter.Get("csv"),", time:",int64(time.Since(start).Seconds()*1e3),"ms")
    return converted, err
}




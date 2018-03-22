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
    //"os"
    "os/exec"
    "sort"
    "strconv"
    "sync"
    "time"
    "gopkg.in/yaml.v2"
    //"github.com/golang/geo"
)


const (
    BaseUrl = "http://localhost:8000"
    ListeningPort = "8000"
    apilog = "./apilog"
)


type Requests struct {
    Shp int64 `json:"shp"`
    Csv int64 `json:"csv"`
    Dem int64 `json:"dem"`
    Dxf int64 `json:"dxf"`
}


type single struct {
    mu     sync.Mutex
    values map[string]int64
}


var counter = single{
    values: make(map[string]int64),
}


func check(e error) bool{
    if e != nil {
      log.Println(e)
      return false
    }
    return true
}


func main() {
    go readCount()

    m := http.NewServeMux()

    proxy := &http.Server{
      Addr:":"+ListeningPort,
      Handler: m,
      MaxHeaderBytes: 30000000,
      ReadTimeout: 10 * time.Second,
    }

    m.HandleFunc("/data", dataHandler)
    m.HandleFunc("/dem", demHandler)
    log.Println("Listening on " + ListeningPort)
    proxy.ListenAndServe()
}


func nullHandler(w http.ResponseWriter, r *http.Request) {
    http.NotFound(w, r)
}


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


func demHandler(w http.ResponseWriter, r *http.Request) {
    start := time.Now()

    ymldata, err := ioutil.ReadAll(r.Body)
    check(err)

    var project Projects
    var hashes []string
    var hash string

    err = yaml.Unmarshal(ymldata, &project)
    check(err)

    for _, dataset := range project.Datasets {
      if len(dataset.S2hash) > 0 {
        hashes = append(hashes, dataset.S2hash)
      }
    }

    sort.Strings(hashes)
    hash = hashes[0]

    out, err := exec.Command(Getdem, hash).Output()
    check(err)

    dem := bytes.NewReader(out)

    io.Copy(w, dem)
    counter.Incr("dem")
    log.Println("dems processed",counter.Get("dem"),", time:",int64(time.Since(start).Seconds()),"s")
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
              case "X": point.X, _ = strconv.Atoi(value)
              case "Y": point.Y, _ = strconv.Atoi(value)
              case "Z": point.Z, _ = strconv.Atoi(value)
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


func (s *single) Get(key string) int64 {
    s.mu.Lock()
    defer s.mu.Unlock()
    return s.values[key]
}

func (s *single) Set(key string, newValue int64) int64 {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.values[key] = newValue
    return s.values[key]
}

func (s *single) Incr(key string) int64 {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.values[key]++
    return s.values[key]
}

func readCount() {
    read, err := ioutil.ReadFile(apilog)
    if err != nil {
      log.Println(err)
      return
    }

    count := Requests{}
    jerr := json.Unmarshal([]byte(read), &count)

    if jerr != nil {
      log.Println(err)
      counter.Set("shp",0)
      counter.Set("dem",0)
      counter.Set("csv",0)
      counter.Set("dxf",0)
      writeCount()
    }

    counter.Set("shp",count.Shp)
    counter.Set("dem",count.Dem)
    counter.Set("csv",count.Csv)
    counter.Set("dxf",count.Dxf)
    log.Println("starting shp count:",counter.Get("shp"))
    log.Println("starting dem count:",counter.Get("dem"))
    log.Println("starting csv count:",counter.Get("csv"))
    log.Println("starting dxf count:",counter.Get("dxf"))

    expiryTime := int64(6)

    writeTicker := time.NewTicker(time.Second * time.Duration(expiryTime))

    for {
      select {
        case <- writeTicker.C:
                writeCount()
      }
    }
}


func writeCount() {
    count := Requests{}
    count.Shp = counter.Get("shp")
    count.Dem = counter.Get("dem")
    count.Csv = counter.Get("csv")
    count.Dxf = counter.Get("dxf")

    writeMe, err := json.Marshal(count)
    if err != nil {
      log.Println(err)
    }

    jerr := ioutil.WriteFile(apilog, writeMe, 0644)
    if jerr != nil {
      log.Println(err)
    }
}

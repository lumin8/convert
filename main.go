package main

import (
    "bytes"
    "encoding/csv"
    "encoding/json"
    "io"
    "io/ioutil"
    "log"
    "net/http"
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


func check(e error) {
    if e != nil {
      log.Println(e)
    }
}


func main() {
    go readCount()

    m := http.NewServeMux()

    proxy := &http.Server{
      Addr:":"+ListeningPort,
      Handler: m,
      ReadTimeout: 0 * time.Second,
    }

    m.HandleFunc("/data/", dataHandler)
    m.HandleFunc("/dem/", demHandler)
    log.Println("Listening on " + ListeningPort)
    proxy.ListenAndServe()
}


func nullHandler(w http.ResponseWriter, r *http.Request) {
    http.NotFound(w, r)
}


func dataHandler(w http.ResponseWriter, r *http.Request) {
    var indataset Datasets
    var converted []byte

    start := time.Now()
    var data []byte
    var info []byte

    reader, err := r.MultipartReader()
    check(err)

    for {
      part, err := reader.NextPart()
      if err == io.EOF {
        break
      }

      switch part.FormName() {
        case "info" :
          info, err = ioutil.ReadAll(part)
          check(err)
        case "file" :
          data, err = ioutil.ReadAll(part)
          check(err)
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
        //outdataset, err = ShpHandler(indataset, contents)
      //case "dxf": 
        //outdataset, err = DxfHandler(indataset, contents)
      default :
        converted = []byte("Sorry, things didn't work out.  Is the format supported?")
    }

    log.Printf("%s\n", converted)
    log.Println("total dataset round trip was ",int64(time.Since(start).Seconds()*1e3),"ms")
    w.Write(converted)
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
    log.Println("dem count",counter.Get("dem"),"ms",int64(time.Since(start).Seconds()*1e3))
}


func CsvHandler(indataset Datasets, contents []byte) (converted []byte, err error) {
    start := time.Now()
    s := bytes.NewReader(contents)

    raw, err := csv.NewReader(s).ReadAll()
    check(err)

    xfield := indataset.Xfield
    yfield := indataset.Yfield
    zfield := indataset.Zfield

    var outdataset Datasets
    var point Points
    var headers map[int]string
    headers = make(map[int]string)

    var attributes Attributes

    for i, record := range raw {
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
    log.Println("dataset count",counter.Get("csv"),"ms",int64(time.Since(start).Seconds()*1e3))
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
    counter.Set("dxf",count.Csv)
    log.Println("starting shp count:",counter.Get("shp"))
    log.Println("starting dem count:",counter.Get("dem"))
    log.Println("starting csv count:",counter.Get("csv"))
    log.Println("starting dxf count:",counter.Get("dxf"))

    expiryTime := int64(600)

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

package main

import (
    "bytes"
    "net/http"
    "sync"
    "io/ioutil"
    "io"
    "log"
    "encoding/json"
    "time"
    "os/exec"
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
    start := time.Now()

    ymldata, err := ioutil.ReadAll(r.Body)
    check(err)

    var dataset Datasets
    var format string

    err = yaml.Unmarshal(ymldata, &dataset)
    check(err)

    thingy := dataset.Url
    log.Println("url: ",thingy)

    for _, info := range Datasets {
      if len(info.Format) > 0 {
        format = info.Format
        log.Println("format: ",info.Format)
      }
    }

    out, err := exec.Command(Convert, format).Output()
    check(err)

    converted := bytes.NewReader(out)

    io.Copy(w, converted)
    log.Println("dataset count",counter.Get("csv"),"ms",int64(time.Since(start).Seconds()*1e3))
}


func demHandler(w http.ResponseWriter, r *http.Request) {
    start := time.Now()

    ymldata, err := ioutil.ReadAll(r.Body)
    check(err)

    var project Project
    var s2hash string

    err = yaml.Unmarshal(ymldata, &project)
    check(err)

    for _, process := range project.Datasets {
      if len(process.S2hash) > 0 {
        s2hash = process.S2hash
        log.Println("format: ",process.S2hash)
      }
    }

    out, err := exec.Command(Getdem, s2hash).Output()
    check(err)

    dem := bytes.NewReader(out)

    io.Copy(w, dem)
    log.Println("dem count",counter.Get("dem"),"ms",int64(time.Since(start).Seconds()*1e3))
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
      counter.Set("shp",1)
      counter.Set("dem",1)
      counter.Set("csv",1)
      counter.Set("dxf",1)
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

package main

import (
    "encoding/json"
    "io/ioutil"
    "log"
    "net/http"
    "sync"
    "time"
)


const (
    BaseUrl = "http://localhost:8000"
    ListeningPort = "8000"
    apilog = "../apilog"
)


type ErrorString struct {
    s string
}


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


func (e *ErrorString) Error() string {
      return e.s
}


func New(text string) error {
      return &ErrorString{text}
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


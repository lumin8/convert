package main

import (
    "net/http"
    "strings"
    "sync"
    "strconv"
    "io/ioutil"
    "io"
    "log"
    "encoding/json"
    "time"
    "testing"
    "math"
    "os"
    "path/filepath"
    "gopkg.in/yaml.v2"
    "github.com/golang/geo"
)


func TestSum(t *testing.T) {
    total := Sum(5, 5)
    if total != 10 {
       t.Errorf("Sum was incorrect, got: %d, want: %d.", total, 10)
    }
}

const (
    BaseUrl = "http://localhost:8000"
    ScriptDir = ""
    BaseDir = "/data/"
    DemDir = "/"
    ListeningPort = "8000"
    Log = "./log"
    csvConv = "./csv2json.py"
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
      log.Println(err)
    }
}


func main() {
    BaseDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))

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

    s := strings.Split(r, "/")
    process := s[1]

    c := exec.Command(process, args)
    out, err := c.Output()
    check(err)

    io.Copy(w, out)
    log.Println("dataset count",counter.Get(process),"ms",int64(time.Since(start).Seconds()*1e3))
}


func demHandler(w http.ResponseWriter, r *http.Request) {
    start := time.Now()

    s := strings.Split(r, "/")
    process := s[1]

    c := exec.Command(process, args)
    out, err := c.Output()
    check(err)

    io.Copy(w, out)
    log.Println("dem count",counter.Get(process),"ms",int64(time.Since(start).Seconds()*1e3))
}


func readCount() {
    read, err := ioutil.ReadFile(log)
    if err != nil {
      log.Println(err)
      return
    }
    count := tileCount{}
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
    count := tileCount{}
    count.Shp = counter.Get("shp")
    count.Dem = counter.Get("dem")
    count.Csv = counter.Get("csv")
    count.Dxf = counter.Get("dxf")

    writeMe, err := json.Marshal(count)
    if err != nil {
      log.Println(err)
    }

    jerr := ioutil.WriteFile(Log, writeMe, 0644)
    if jerr != nil {
      log.Println(err)
    }
}

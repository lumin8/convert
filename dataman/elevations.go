package dataman


import (
     "bytes"
     "gopkg.in/yaml.v2"
     "io/ioutil"
     "os"
)


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


func getElev(x, y, z, elev int) int {
    if value, ok := os.LookupEnv(z); ok {
        return value
    }

    cmd="gdallocationinfo -xml " + usadem + " -geoloc " + x + " " + y

    out, err := exec.Command("sh", "-c", cmd).Output()
    return fallback
}

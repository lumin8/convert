package main


import (
     "bufio"
     "encoding/json"
     "errors"
     "log"
     "net/http"
     "os"
     "os/exec"
     "strconv"
     "strings"
     "time"
)


const (
    demvrt = "/data/raw/dem/hdt/earthdem.vrt"
)


func paramCheck(i string, r *http.Request) (string, []byte) {
    log.Println("paramcheck...")
    val, ok := r.URL.Query()[i]

    log.Println("var lookedup...")
    var resp []byte
    if !ok || len(val) < 1 {
        resp  = []byte("Please provide valid x and y parameters in lat/long decimal degrees.")
        return "", resp
    }

    log.Println("checking string...")

    str := val[0]

    return str, resp
}


func demHandler(w http.ResponseWriter, r *http.Request) {
    start := time.Now()

    x, resp := paramCheck("x", r)
    if resp != nil {
        w.Write(resp)
    }

    y, resp := paramCheck("y", r)
    if resp != nil {
        w.Write(resp)
    }

    format, resp := paramCheck("f", r)
    if resp != nil {
        format = "json"
    }

    data, err := getDem(x,y)
    check(err)

    //TBD hash support
    /***hashes := r.URL.Query()["hash"]
    var project Projects
    var hashes []string
    var hash string
    for _, dataset := range project.Datasets {
      if len(dataset.S2hash) > 0 {
        hashes = append(hashes, dataset.S2hash)
      }
    }
    sort.Strings(hashes)
    hash = hashes[0] 
    data, err := getDem(hash) ...***/

    switch format {
      case "json" :
        out, _ := json.Marshal(data)
        w.Write(out)
      case "struct" :
        //out := bytes.NewReader(data)
        //w.Write(data)
    }

    counter.Incr("dem")
    log.Println(counter.Get("dem"),"dems processed, time:",int64(time.Since(start).Seconds()*1e3),"ms")
}


func getElev(x string, y string) (string, error) {
    var zstr string

    if _, err := os.Stat(demvrt); err !=  nil {
      err = errors.New("Sorry, the world digital elevation model (DEM) is unavailable")
      return zstr, err
    }

    cmd := "gdallocationinfo -valonly " + demvrt + " -geoloc " + x + " " + y
    z, err := exec.Command("sh", "-c", cmd).Output()
    zstr = string(z)
    check(err)
    return zstr, err
}


func getDem(x string, y string) (data Dem, err error) {
    start := time.Now()

    var dem Dem

    if _, err := os.Stat(demvrt); err !=  nil {
      err = errors.New("Sorry, the world digital elevation model (DEM) is unavailable")
      return dem, err
    }

    //TBD make this more elegant, remove dep on writing files to disk
    xint, _ := strconv.ParseFloat(x, 64)
    yint, _ := strconv.ParseFloat(y, 64)
    uly := strconv.FormatFloat(yint + 0.03, 'f', -2, 64)
    ulx := strconv.FormatFloat(xint - 0.03, 'f', -2, 64)
    bry := strconv.FormatFloat(yint - 0.03, 'f', -2, 64)
    brx := strconv.FormatFloat(xint + 0.03, 'f', -2, 64)

    tif4326 := "mktemp " + basepath + "/" + "../tmp/XXXXXX.tif"
    tif3857 := "mktemp " + basepath + "/" + "../tmp/XXXXXX.tif"
    xyz3857 := "mktemp " + basepath + "/" + "../tmp/XXXXXX.xyz"

    mktifin, _ := exec.Command("sh", "-c", tif4326).Output()
    mktifout, _ := exec.Command("sh", "-c", tif3857).Output()
    mkxyz, _ := exec.Command("sh", "-c", xyz3857).Output()

    tifin := strings.TrimSpace(string(mktifin))
    tifout := strings.TrimSpace(string(mktifout))
    xyzfile := strings.TrimSpace(string(mkxyz))

    clip := "gdal_translate -projwin " + ulx + " " + uly + " " + brx + " " + bry + " " + demvrt + " " + tifin
    transform := "gdalwarp -s_srs EPSG:4326 -t_srs EPSG:3857 " + tifin + " " + tifout
    convert := "gdal_translate -of XYZ " + tifout + " " + xyzfile

    cmds := clip + "; " + transform + "; " + convert

    exec.Command("sh", "-c", cmds).Run()

    os.Remove(tifin)
    os.Remove(tifout)

    xyz, err := os.Open(xyzfile)
    check(err)
    defer xyz.Close()

    scanner := bufio.NewScanner(xyz)
    for scanner.Scan() {
      //var point Point
      values := strings.Fields(scanner.Text())
      X := str2fixed(values[0])
      Y := str2fixed(values[1])
      Z := str2fixed(values[2])
      point := []float64{X,Y,Z}
      dem.Points = append(dem.Points, point)
    }

    xyz.Close()
    os.Remove(xyzfile)

    log.Println("total dataset round trip:",int64(time.Since(start).Seconds()*1e3),"ms")
    return dem, err
}


func str2fixed(num string) float64 {
    val, _ := strconv.ParseFloat(num,64)
    j := strconv.FormatFloat(val, 'f', 2, 64)
    i, _ := strconv.ParseFloat(j,64)
    return i
}

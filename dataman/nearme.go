package main

import (
    "database/sql"
    "encoding/json"
    "errors"
    "io/ioutil"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/user"
    "sync"
    "time"
    _ "github.com/lib/pq"
)

const (
    dbname = "osm"

    // table names
    trails = "trails"
    points = "poi"
    shapes = "planet_osm_polygon"
    roads = "roads"
    rivers = "rivers"

    // style file
    styles = "config/styles.json"

    pgcreds = "/.pgpass"
)


func nearmeHandler(w http.ResponseWriter, r *http.Request) {
    start := time.Now()

    flavor, err := paramCheck("styles", r)
    if err != nil {
        flavor = ""
    } else {
        f, err := ioutil.ReadFile(styles)
        if err != nil {
          log.Printf("%s",err)
          w.Write([]byte("could not fetch stylesheet"))
        }
        log.Println("stylesheet request")
        log.Printf("%s",f)
        w.Write(f)
        r.Body.Close()
        return
    }

    x, err := paramCheck("x", r)
    if err != nil {
        w.Write(err)
        r.Body.Close()
        return
    }

    y, err := paramCheck("y", r)
    if err != nil {
        w.Write(err)
        r.Body.Close()
        return
    }

    format, err := paramCheck("f", r)
    if err != nil {
        format = "json"
    }

    flavor, err = paramCheck("type", r)
    if err != nil {
        flavor = "poi"
    }

    data, resp := fetchData(x,y,flavor)
    if resp != nil {
        check(resp)
    }

    switch format {
      case "json" :
        out, _ := json.Marshal(data)
        w.Write(out)
      case "struct" :
        //out := bytes.NewReader(data)
        //w.Write(data)
    }

    r.Body.Close()

    counter.Incr("dem")
    log.Println(counter.Get("dem"),"dems processed, time:",int64(time.Since(start).Seconds()*1e3),"ms")
}


func fetchData(x string, y string, flavor string) (Datasets, error) {
    var table string
    var meters string
    var data Datasets

    log.Println("nearme query for",flavor,"around",x,y)

    usr, err := user.Current()
    check(err)
    pgpass := usr.HomeDir + pgcreds

    if _, err := os.Stat(pgpass); err !=  nil {
      cerr := errors.New("Missing or misconfigured credentials pgpass specified in the host's home directory.")
      return data, cerr
    }

    dbinfo := fmt.Sprintf("dbname=%s sslmode=disable", dbname)
    db, err := sql.Open("postgres", dbinfo)
    if err != nil {
      cerr := errors.New("Could not establish a connection with the host")
      return data, cerr
    }
    defer db.Close()

    err = db.Ping()
    if err != nil {
      cerr := errors.New("Could not establish a connection with the dataset")
      return data, cerr
    }

    switch flavor {
      case "trails" :
        table = trails
        meters = "5000"
      case "roads" :
        table = roads
        meters = "500"
      case "shapes" :
        table = shapes
        meters = "1000"
      case "poi" :
        table = points
        meters = "10000"
      default :
        errors.New("Either no data exists, or your request is not supported")
        return data, err
    }

    query := "with nearme as (select name,way FROM " + table + " WHERE ST_Intersects(way, ST_Buffer(ST_Transform(ST_SetSRID(ST_MakePoint("+ x +", "+ y +"), 4326), 900913), "+ meters +")) and name is not null and way is not null) select name, st_asgeojson(ST_Intersection(way, ST_Buffer(ST_Transform(ST_SetSRID(ST_MakePoint("+ x +", "+ y +"), 4326), 900913), "+ meters +")) ) from nearme "

//    query := "with nearme as (select name,way FROM " + table + " WHERE ST_DWithin(way, ST_Transform(ST_SetSRID(ST_MakePoint("+ x +", "+ y +"), 4326), 900913), "+ meters +") and name is not null and way is not null) select name, st_asgeojson(st_intersection(way,ST_Buffer(ST_Transform(ST_SetSRID(ST_MakePoint("+ x +", "+ y +"), 4326), 900913), "+ meters +"))) from nearme "

    rows, err := db.Query(query)
    defer rows.Close()

    if err == sql.ErrNoRows {
      cerr := errors.New("No Results Found")
      return data, cerr
    }

    if err != nil {
      log.Printf("%s",err)
      return data, err
    }

    var wg sync.WaitGroup

    for rows.Next() {

      var name string
      var geom string
      err = rows.Scan(&name, &geom)
      if err != nil {
        continue
      }
      //log.Println(name,geom)

      wg.Add(1)

      go func() {
        switch table {
           case "points", "poi", "pois":
             var feature Points
             var attributes Attributes
             feature.Point = derivePoint(geom).Points[0]
             attributes.Key = "name"
             attributes.Value = name
             feature.Attributes = append(feature.Attributes, attributes)
             data.Points = append(data.Points, feature)
           case "trails", "roads", "rivers":
             var feature Lines
             feature.Name = name
             feature.Points = derivePoints(geom).Points
             data.Lines = append(data.Lines, feature)
           case shapes:
             var feature Shapes
             feature.Name = name
             feature.Points = derivePoints(geom).Points
             data.Shapes = append(data.Shapes, feature)
        }
        wg.Done()
      }()

    }

    wg.Wait()

    return data, err
}


func derivePoints(geom string) Pointarray {

    var geojson GeojsonM
    var coords Pointarray
    err := json.Unmarshal([]byte(geom), &geojson)
    if err != nil {
      return coords
    }

    for _, point := range geojson.Coords {
      var z float64
      if len(point) < 3 {
        z, err = getElev(point[0],point[1])
        if err != nil {
          log.Printf("%s",err)
          return coords
        }
      }
      var xyz []float64
      xyz = append(xyz, point[0], point[1], z)
      coords.Points = append(coords.Points, xyz)
    }

    return coords
}

func derivePoint(geom string) Pointarray {
    var geojson GeojsonS
    var coords Pointarray

    err := json.Unmarshal([]byte(geom), &geojson)
    if err != nil {
      return coords
    }

    if len(geojson.Coords) < 3 {
      z, err := getElev(geojson.Coords[0],geojson.Coords[1])
      if err != nil {
        log.Printf("%s",err)
      }
      geojson.Coords = append(geojson.Coords, z)
    }
    coords.Points = append(coords.Points, geojson.Coords)

    return coords
}

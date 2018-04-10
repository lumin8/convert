package main

import (
    "database/sql"
    "encoding/json"
    "errors"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"
    _ "github.com/lib/pq"
)

const (
    dbname = "osm"
    lines = "planet_osm_lines"
    points = "planet_osm_point"
    shapes = "planet_osm_polygon"
    ways = "planet_osm_roads"
    pgpass = "$HOME/.pgpass"
)


func nearmeHandler(w http.ResponseWriter, r *http.Request) {
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

    flavor, resp := paramCheck("type", r)
    if resp != nil {
        flavor = "poi"
    }

    data, err := fetchData(x,y,flavor)
    check(err)

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


func fetchData(x string, y string, flavor string) (Datasets, error) {
    var table string
    var wheresql string
    var meters string
    var data Datasets

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
        table = lines
        wheresql = "type ILIKE 'trail' or type ILIKE '%path%'"
        meters = "20000"
      case "roads" :
        table = ways
        wheresql = "type ILIKE '%roads%'"
        meters = "2000"
      case "shapes" :
        table = shapes
        wheresql = ""
        meters = "2000"
      case "poi" :
        table = points
        wheresql = ""
        meters = "2000"
      default :
        errors.New("Either no data exists, or your request is not supported")
        return data, err
    }

    query := "SELECT name, st_asgeojson(way), FROM " + table + " " + wheresql
    query = query + "and WHERE ST_Within(the_geom, ST_Transform(ST_Buffer(ST_Transform(ST_SetSRID(ST_MakePoint(" + x + ", " + y + "), 4326), 3857), " + meters + "), 4326)) = 1"
    rows, err := db.Query(query)

    if err == sql.ErrNoRows {
      cerr := errors.New("No Results Found")
      return data, cerr
    }

    if err != nil {
      return data, err
    }

    for rows.Next() {
      var name string
      var geom string
      err = rows.Scan(&name, &geom)
      check(err)
      switch table {
         case "points":
           var feature Points
           var attributes Attributes
           feature.Points = derivePoints(geom).Points
           attributes.Key = "name"
           attributes.Value = name
           feature.Attributes = append(feature.Attributes, attributes)
           data.Points = append(data.Points, feature)
         case "lines", "ways":
           var feature Lines
           feature.Name = name
           feature.Points = derivePoints(geom).Points
           data.Lines = append(data.Lines, feature)
         case "shapes":
           var feature Shapes
           feature.Name = name
           feature.Points = derivePoints(geom).Points
           data.Shapes = append(data.Shapes, feature)
      }
    }
    return data, err
}


func derivePoints(geom string) Pointarray {
    var geojson Geojson
    var jsonBlob = []byte(geom)
    var coords Pointarray
    err := json.Unmarshal(jsonBlob, &geojson)
    check(err)

    for i, record := range geojson.Coords {
      if len(record) < 3 {
        z, err := getElev(record[0],record[1])
        if err != nil {
          log.Printf("%s",err)
          continue
        }
        geojson.Coords[i] = append(geojson.Coords[i], z)
      }
      coords.Points = append(coords.Points, record)
    }

    return coords
}

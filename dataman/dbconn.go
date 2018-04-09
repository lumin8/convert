package main

import (
    "database/sql"
    "errors"
    "fmt"
    "os"
    "gopkg.in/yaml.v2"
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
           feature.Point = derivePoints(geom)
           attributes.Key = "name"
           attributes.Value = name
           feature.Attributes = append(feature.Attributes, attributes)
           data.Points = append(data.Points, feature)
         case "lines", "ways":
           var feature Lines
           feature.Name = name
           feature.Points = derivePoints(geom)
           data.Lines = append(data.Lines, feature)
         case "shapes":
           var feature Shapes
           feature.Name = name
           feature.Points = derivePoints(geom)
           data.Shapes = append(data.Shapes, feature)
      }
    }
    return data, err
}


func derivePoints(geom string, feature string) Coords {
    var geojson Geojson
    data := yaml.Unmarshal(geom, &geojson)
    for i, record := range data.Coords {
      if len(record) < 3 {
        z, err := getElev(x,y)
        if err != nil {
          log.Printf("%s",err)
          continue
        }
        data.Coords[i] = append(data.Coords[i], z)
      }
    }
    return data.Coords
}

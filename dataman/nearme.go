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
    "github.com/paulmach/go.geojson"
    _ "github.com/lib/pq"
    "github.com/remeh/sizedwaitgroup"
    "github.com/golang/geo/s2"
)

const (
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
        w.Write(f)
        r.Body.Close()
        return
    }

    x, err := paramCheck("x", r)
    if err != nil {
        log.Println(err)
    }

    y, err := paramCheck("y", r)
    if err != nil {
        log.Println(err)
    }

    s2key, err := paramCheck("s2", r)
    if err != nil {
        log.Println(err)
    }

    if len(s2key) > 0 {
      x,y = s2xy(s2key)
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
    var meters string
    var data Datasets
    dbname := "osm"

    // these are for the postgres parsing
    var geomcol, gjson, geomtype, srid string

    log.Println("nearme query for",flavor,"around",x,y)

    usr, err := user.Current()
    check(err)
    pgpass := usr.HomeDir + pgcreds

    if _, err := os.Stat(pgpass); err !=  nil {
      cerr := errors.New("Missing or misconfigured credentials pgpass specified in the host's home directory.")
      return data, cerr
    }

    var processes = 150 // number concurrent threads per feature flavor

    switch flavor {
      case "trails", "rivers" :
        meters = "1500"
      case "roads" :
        meters = "1000"
      case "shapes" :
        meters = "1000"
      case "poi" :
        meters = "10000"
        processes = 50
      default :
        //errors.New("Either no data exists, or your request is not supported")
        meters = "10000"
        dbname = "sample"
        //return data, err
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


    // get column name and type of geom (point, linestring, polygon)
    query := "select f_geometry_column as name,type,srid from geometry_columns where f_table_name = '" + flavor + "';"

    rows, err := db.Query(query)
    defer rows.Close()
    if err == errors.New("No Results Found") {
      log.Printf("%s",err)
    }

    if err != nil {
      log.Printf("%s",err)
    }

    for rows.Next() {
      err = rows.Scan(&geomcol, &geomtype, &srid)
      if err != nil {
        log.Printf("%s",err)
      }
    }

    rows.Close()

    switch dbname {
      case "osm":
        query = "select st_asgeojson(ST_Intersection("+ geomcol +
        ", ST_Buffer(ST_Transform(ST_SetSRID(ST_MakePoint("+ x +", "+ y +
        "), 4326), 900913), "+ meters +")) ), "+
        "jsonb_build_object(" +
        "'type','FeatureInfo'," +
        "'properties',to_jsonb(row) - 'osm_id' - 'tags' - '"+geomcol+"')" +
        " FROM (SELECT * FROM " + flavor +
        " where ST_Intersects(way, ST_Buffer(ST_Transform(ST_SetSRID(ST_MakePoint("+
        x +", "+ y +"), 4326), 900913), "+ meters +"))) row;"
      default:
        query = "select st_asgeojson("+ geomcol +
        "),jsonb_build_object(" +
        "'type','FeatureInfo'," +
        "'properties',to_jsonb(row) - 'osm_id' - 'tags' - '"+geomcol+"')" +
        " FROM (SELECT * FROM " + flavor + ") row;"
        log.Println("query: ",query)
    }

    rows, err = db.Query(query)
    defer rows.Close()

    if err == sql.ErrNoRows {
      cerr := errors.New("No Results Found")
      return data, cerr
    }

    if err != nil {
      log.Printf("%s",err)
      return data, err
    }

    wg := sizedwaitgroup.New(processes)
    for rows.Next() {

      var geom string
      var pgfeature FeatureInfo

      err = rows.Scan(&geom,&gjson)
      if err != nil {
        log.Printf("%s",err)
        return data, err
      }

      wg.Add()

      go func() {

      defer wg.Done()

      err = json.Unmarshal([]byte(gjson), &pgfeature.geojson)
      if err != nil {
        log.Printf("%s",err)
        return
      }

      pgfeature.geojson.Geometry, err = geojson.UnmarshalGeometry([]byte(geom))
      if err != nil {
        log.Printf("%s",err)
        return
      }

      switch pgfeature.geojson.Geometry.Type {
        case "Point":
          var wg sync.WaitGroup
          var feature Points
          wg.Add(2)
          go func() {
            defer wg.Done()
            feature.Point = derivePoints(&pgfeature).Points[0]
            if len(feature.Point) < 1 {
              wg.Done()
              return
            }
          }()
          go func() {
            defer wg.Done()
            feature.Attributes = parseAttributes(&pgfeature)
            feature.Name = pgfeature.name
            feature.StyleType = pgfeature.styletype
          }()
          wg.Wait()
          data.Points = append(data.Points, feature)

        case "LineString":
          var wg sync.WaitGroup
          var feature Lines
          wg.Add(2)
          go func() {
            defer wg.Done()
            feature.Attributes = parseAttributes(&pgfeature)
            feature.Name = pgfeature.name
            feature.StyleType = pgfeature.styletype
          }()
          go func() {
            defer wg.Done()
            feature.Points = derivePoints(&pgfeature).Points
            if len(feature.Points) < 1 {
              wg.Done()
              return
            }
          }()
          wg.Wait()
          data.Lines = append(data.Lines, feature)

        case "Polygon":
          var wg sync.WaitGroup
          var feature Shapes
          wg.Add(2)
          go func() {
            defer wg.Done()
            feature.Attributes = parseAttributes(&pgfeature)
            feature.Name = pgfeature.name
            feature.StyleType = pgfeature.styletype
          }()
          go func() {
            defer wg.Done()
            feature.Points = derivePoints(&pgfeature).Points
            if len(feature.Points) < 1 {
              wg.Done()
              return
            }
          }()
          wg.Wait()
          data.Shapes = append(data.Shapes, feature)
      }

      }()
    }

    wg.Wait()

    return data, err
}


func parseAttributes (pgfeature *FeatureInfo) []Attributes { //[]map[string]interface{} {
    // TBD geojson var atts []map[string]interface{}
    var atts []Attributes
    for k, v := range pgfeature.geojson.Properties {
      switch v {
        case nil, "", 0, "0":
          delete(pgfeature.geojson.Properties, k)
          continue
      }
      switch k {
        case "name":
          pgfeature.name = fmt.Sprintf("%v",v)
        case "styletype":
          pgfeature.styletype = fmt.Sprintf("%v",v)
        default:
            var attrib Attributes
            attrib.Key = k
            attrib.Value = fmt.Sprintf("%v",v)
            atts = append(atts, attrib)

            //TBD geojson pair := make(map[string]interface{})
            //TBD geojson pair[k] = v
            //TBD geojson atts = append(atts, pair)
      }
    }
    return atts
}


func derivePoints(pgfeature *FeatureInfo) Pointarray {
    var coordarray Pointarray

    switch pgfeature.geojson.Geometry.Type {

    case "Point" :
      var z float64
      x, y := To3857(pgfeature.geojson.Geometry.Point[0], pgfeature.geojson.Geometry.Point[1])
      if (x == 0 && y == 0) {
        return coordarray
      }
      if len(pgfeature.geojson.Geometry.Point) < 3 {
         z, _ = getElev(x,y)
      }

      coordarray.Points = append(coordarray.Points, []float64{x,y,z})

    case "LineString" :
      var z float64
      for _, coords := range pgfeature.geojson.Geometry.LineString {
        x, y := To3857(coords[0], coords[1])
        if (x == 0 && y == 0) {
          continue
        }
        if len(coords) < 3 {
          z, _ = getElev(x,y)
        }
        coordarray.Points = append(coordarray.Points, []float64{x,y,z})
      }

    case "Polygon" :
      var z float64
      for _, coords := range pgfeature.geojson.Geometry.Polygon {
        for _, coord := range coords {
            x, y := To3857(coord[0], coord[1])
            if (x == 0 && y == 0) {
              continue
            }
            if len(coord) < 3 {
              z, _ = getElev(x,y)
            }

            coordarray.Points = append(coordarray.Points, []float64{x,y,z})
        }
      }

    }
    return coordarray
}


func s2xy (token string) (string, string) {
    s2cellid := s2.CellIDFromToken(token)
    s2cell := s2.CellFromCellID (s2cellid)
    //Scott TBD add area param
    //area := s2cell.ApproxArea()
    point := s2cell.Center()
    ll := s2.LatLngFromPoint(point)
    xa := ll.Lng //Longitude()
    ya := ll.Lat //itude()
    xd := xa.Degrees()
    yd := ya.Degrees()
    x := fmt.Sprintf("%v",xd)
    y := fmt.Sprintf("%v",yd)
    return x, y
}

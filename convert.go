package convert

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/paulmach/go.geo"
	"github.com/golang/geo/s2"
)

// Datasets ...
type Datasets struct {
	Id      string       `json:"id" yaml:"id"`
	Name    string       `json:"name" yaml:"name"`
	Url     string       `json:"dataurl" yaml:"dataurl"`
	Updated string       `json:"lastUpdated" yaml:"lastUpdated"`
	Center  []Coordinate `json:"center" yaml:"center"`
	S2      []string     `json:"s2" yaml:"s2"`
	Points  []Point      `json:"points" yaml:"points"`
	Lines   []Line       `json:"lines" yaml:"lines"`
	Shapes  []Shape      `json:"shapes" yaml:"shapes"`
}

// Coordinate ...
type Coordinate struct {
	X float64 `json:"x" yaml:"x"`
	Y float64 `json:"y" yaml:"y"`
	Z float64 `json:"z" yaml:"z"`
}

// Point ...
type Point struct {
	Id         string      `json:"id" yaml:"id"`
	Name       string      `json:"name" yaml:"name"`
	StyleType  string      `json:"type" yaml:"type"`
	Attributes []Attribute `json:"attributes" yaml:"attributes"`
	Point      []float64   `json:"point" yaml:"point"`
}

// Points ...
type PointArray struct {
        Points []([]float64) `json:"points" yaml:"points"`
}

// Line ...
type Line struct {
	Id         string      `json:"id" yaml:"id"`
	Name       string      `json:"name" yaml:"name"`
	StyleType  string      `json:"type" yaml:"type"`
	Attributes []Attribute `json:"attributes" yaml:"attributes"`
	Points     [][]float64 `json:"points" yaml:"points"`
}

// Shape ...
type Shape struct {
	Id         string      `json:"id" yaml:"id"`
	Name       string      `json:"name" yaml:"name"`
	StyleType  string      `json:"type" yaml:"type"`
	Attributes []Attribute `json:"attributes" yaml:"attributes"`
	Points     [][]float64 `json:"points" yaml:"points"`
}

// Attribute ...
type Attribute struct {
	Key   string `json:"key" yaml:"key"`
	Value string `json:"value" yaml:"value"`
}

// GeojsonS ...
type GeojsonS struct {
	Type   string    `json:"type" yaml:"type"`
	Coords []float64 `json:"coordinates" yaml:"coordinates"`
}

// GeojsonM ...
type GeojsonM struct {
	Type   string      `json:"type" yaml:"type"`
	Coords [][]float64 `json:"coordinates" yaml:"coordinates"`
}

// FeatureInfo ...
type FeatureInfo struct {
	id        int
//	geojson   geojson.Feature //G
	geomtype  string
	srid      string
	s2        []s2.CellID
	tokens    []string
	name      string
	styletype string
}

const (
	// env var for the dem.ver path
	envDEMVRT = "DEMVRT"

	// process limits for the sized wait group
	50
)

// demvrt is used to cache the path of the dem.vrt file after it has been resolved once.
// Note: if the file is moved or deleted the path will not change
var demvrt = ""

// demvrtPath is used to resolve the path for the dem.vrt file
func demvrtPath() (string, error) {
	if demvrt != "" {
		return demvrt, nil
	}

	dvp := os.Getenv(envDEMVRT)
	if len(dvp) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		dvp = path.Join(cwd, "dem.vrt")
	}

	if _, err := os.Stat(dvp); err != nil {
		return "", fmt.Errorf("error: world digital elevation model (DEM) cannot be found at %s", demvrt)
	}
	demvrt = dvp
	return dvp, nil
}

// DatasetFromCSV ...
func DatasetFromCSV(xField string, yField string, zField string, contents io.Reader) (*Datasets, error) {
	raw, err := csv.NewReader(contents).ReadAll()
	if err != nil {
		return nil, err
	}

	var outdataset Datasets

	headers := make(map[int]string)
	bbox := make(map[string]float64)

	ch := make(chan Point)
        go MinMax(bbox, ch)

	var wg sync.WaitGroup

	for i, record := range raw {
		switch i {
		case 0:
			for i, header := range record {
				switch header {
				case xField:
					headers[i] = "X"
				case yField:
					headers[i] = "Y"
				case zField:
					headers[i] = "Z"
				default:
					headers[i] = header
				}
			}
		default:
			wg.Add(1)
			go ParseCSV(bbox, headers, record, &wg, &outdataset)
		}
	}

	wg.Wait()

	// configure the center point... in 4326
	var c Coordinate
	c.X = bbox["rx"] - (bbox["rx"]-bbox["lx"])/2
	c.Y = bbox["uy"] - (bbox["uy"]-bbox["ly"])/2
	c.Z, _ = GetElev(c.X, c.Y)
	outdataset.Center = append(outdataset.Center, c)

	// configure the s2 array... in 4326
	outdataset.S2 = s2covering(bbox)

	return &outdataset, err
}


// DatasetFromGEOJSON ...
func DatasetFromGEOJSON(xField string, yField string, zField string, contents io.Reader) (*Datasets, error) {
        raw, err := strings.NewReader(contents)
        if err != nil {
                return nil, err
        }

        var outdataset Datasets

	raw, err := geojson.UnmarshalFeatureCollection(raw)

        var outdataset Datasets
        ch := make(chan Point)
        bbox := make(map[string]float64)

        go MinMax(bbox, ch)

        // this kicks off the processing of the data... a biggie!
        err = ParseGEOJSON(raw, &outdataset, ch)
        if err != nil {
                return &outdataset, err
        }

        // configure the center point... in 4326
        var c Coordinate
        c.X = bbox["rx"] - (bbox["rx"]-bbox["lx"])/2
        c.Y = bbox["uy"] - (bbox["uy"]-bbox["ly"])/2
        c.Z, _ = getElev(c.X, c.Y)
        outdataset.Center = append(outdataset.Center, c)

        // configure the s2 array... in 4326
        outdataset.S2 = s2covering(bbox)

        // prepare the final UNITY json
        converted, err = json.Marshal(outdataset)

        return &outdataset, err
}


// ParseCSV ...
func ParseCSV (bbox map[string]float64, headers map[int]string, record []string, wg *sync.WaitGroup, outdataset *Datasets) {

	defer wg.Done()

	var pointxyz Coordinate
        var point Point
	var err error

	for i, value := range record {
		switch headers[i] {
		case "X":
			pointxyz.X, _ = strconv.ParseFloat(value, 64)
		case "Y":
			pointxyz.Y, _ = strconv.ParseFloat(value, 64)
		case "Z":
			pointxyz.Z, _ = strconv.ParseFloat(value, 64)

		default:
			var atts Attribute
			atts.Key = headers[i]
			atts.Value = fmt.Sprintf("%v", value)
			//TBD geojson pair := make(map[string]interface{})
			//TBD geojson pair[headers[i]] = value
			point.Attributes = append(point.Attributes, atts)
		}
	}

	// fill elevation if required
	if pointxyz.Z == 0 && pointxyz.X != 0 && pointxyz.Y != 0 {
		log.Printf("value needed filling in with elevation...")
		pointxyz.Z, err = GetElev(pointxyz.X, pointxyz.Y)
		if err != nil {
			log.Printf("couldn't add elevation, reason: %s",err.Error())
		}
	}

	// make SURE X and Y are in 3857
	pointxyz.X, pointxyz.Y = To3857(pointxyz.X, pointxyz.Y)

	// keep a collective of the min / max coords of dataset
        ch <- pointxyz

	// fill in the point float array
	point.Point = append(point.Point, pointxyz.X, pointxyz.Y, pointxyz.Z)

	// finally, append point to the final dataset
	outdataset.Points = append(outdataset.Points, point)
}


// MinMax ...
func MinMax(bbox map[string]float64, ch chan Coordinate) {
	for {

		_, ok := bbox["lx"]
		if !ok {
			bbox["lx"] = X
			bbox["rx"] = X
			bbox["ly"] = Y
			bbox["uy"] = Y
		}

		if X < bbox["lx"] {
			bbox["lx"] = X
		}
		if X > bbox["rx"] {
			bbox["ux"] = X
		}
		if Y < bbox["ly"] {
			bbox["ly"] = Y
		}
		if Y > bbox["uy"] {
			bbox["uy"] = Y
		}
	}
}

// s2covering ...
func s2covering(bbox map[string]float64) []string {
	var s2hash []string

	rx, uy := To4326(bbox["rx"], bbox["uy"])
	lx, ly := To4326(bbox["lx"], bbox["ly"])
	cz, err := GetElev(bbox["rx"], bbox["uy"])
	if err != nil {
		fmt.Println(err)
	}

	pts := []s2.Point{
		s2.PointFromCoords(rx, uy, cz),
		s2.PointFromCoords(lx, uy, cz),
		s2.PointFromCoords(lx, ly, cz),
		s2.PointFromCoords(rx, ly, cz),
	}

	loop := s2.LoopFromPoints(pts)
	covering := loop.CellUnionBound()

	for _, cellid := range covering {
		token := cellid.ToToken()
		if len(token) > 8 {
			runes := []rune(token)
			token = string(runes[0:8])
		}
		for _, b := range s2hash {
			if b != token {
				continue
			}
			s2hash = append(s2hash, token)
		}
	}

	return s2hash
}

// GetElev gets the elevation for the given x y coordinate
func GetElev(x float64, y float64) (float64, error) {
	// outputs in meters, works regardless of input projection
	lon, lat := To4326(x, y)

	var zstr string

	xstr := strconv.FormatFloat(lon, 'f', -2, 64)
	ystr := strconv.FormatFloat(lat, 'f', -2, 64)

	demvrt, err := demvrtPath()
	if err != nil {
		return 0, err
	}

	cmd := "gdallocationinfo -valonly " + demvrt + " -geoloc " + xstr + " " + ystr
	zbyte, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return 0, err
	}
	zstr = strings.TrimSpace(string(zbyte))
	z, _ := strconv.ParseFloat(zstr, 64)
	return z, err
}

// str2fixed ...
func str2fixed(num string) float64 {
	val, _ := strconv.ParseFloat(num, 64)
	j := strconv.FormatFloat(val, 'f', 2, 64)
	k, _ := strconv.ParseFloat(j, 64)
	return k
}

// To4326 ...
func To4326(x float64, y float64) (float64, float64) {
	// regardless of inbound, kicks out 4326
	if (x <= 180) && (x >= -180) {
		return x, y
	}
	mercPoint := geo.NewPoint(x, y)
	geo.Mercator.Inverse(mercPoint)
	return mercPoint[0], mercPoint[1]
}

// To3857 ...
func To3857(x float64, y float64) (float64, float64) {
	// regardless of inbound, kicks out 3857
	if (x < 180) || (x > -180) {
		mercPoint := geo.NewPoint(x, y)
		geo.Mercator.Project(mercPoint)
		x = mercPoint[0]
		y = mercPoint[1]
	}

	x = math.Round(mercPoint[0]*100)/100
        y = math.Round(mercPoint[1]*100)/100
        return x, y
}

//ParseGEOJSON ...
func ParseGEOJSON(Features *geojson.FeatureCollection, outdataset *Datasets, ch chan Point) error {
        wg := sizedwaitgroup.New(processes)
        var err error

        //access each of the individual features of the geojson
        for _, item := range Features.Features {

		// the entire json feature
                temp, _ := json.Marshal(item)

		// just the geometry
                geom, _ := json.Marshal(item.Geometry)

		// the new feature
                var gfeature FeatureInfo

                wg.Add()

                go func() {

                        defer wg.Done()

                        err = json.Unmarshal(temp, &gfeature.geojson)
                        if err != nil {
                                log.Printf("error unmarshaling feature: %s", err)
                                return
                        }

                        gfeature.geojson.Geometry, err = geojson.UnmarshalGeometry([]byte(geom))
                        if err != nil {
                                log.Printf("error unmarshaling feature geom: %s", err)
                                return
                        }

                        switch gfeature.geojson.Geometry.Type {
                        case "Point","PointZ","point","pointz":
                                var wg sync.WaitGroup
                                var feature Points
                                wg.Add(2)
                                go func() {
                                        defer wg.Done()
                                        feature.Point = deriveGEOJSONPoints(&gfeature,ch).Points[0]
                                        if len(feature.Point) < 1 {
                                                wg.Done()
                                                return
                                        }
                                }()
                                go func() {
                                        defer wg.Done()
                                        feature.Attributes = parseGEOJSONAttributes(&gfeature)
                                        feature.Name = gfeature.name
                                        feature.StyleType = gfeature.styletype
                                }()
                                wg.Wait()
                                outdataset.Points = append(outdataset.Points, feature)

                        case "LineString","LineStringZ","linestring","linestringz":
                                var wg sync.WaitGroup
                                var feature Lines
                                wg.Add(2)
                                go func() {
                                        defer wg.Done()
                                        feature.Attributes = parseGEOJSONAttributes(&gfeature)
                                        feature.Name = gfeature.name
                                        feature.StyleType = gfeature.styletype
                                }()
                                go func() {
                                        defer wg.Done()
                                        feature.Points = deriveGEOJSONPoints(&gfeature,ch).Points
                                        if len(feature.Points) < 1 {
                                                wg.Done()
                                                return
                                        }
                                }()
                                wg.Wait()
                                outdataset.Lines = append(outdataset.Lines, feature)

                        case "Polygon","PolygonZ","polygon","polygonz":
                                var wg sync.WaitGroup
                                var feature Shapes
                                wg.Add(2)
                                go func() {
                                        defer wg.Done()
                                        feature.Attributes = parseGEOJSONAttributes(&gfeature)
                                        feature.Name = gfeature.name
                                        feature.StyleType = gfeature.styletype
                                }()
                                go func() {
                                        defer wg.Done()
                                        feature.Points = deriveGEOJSONPoints(&gfeature,ch).Points
                                        if len(feature.Points) < 1 {
                                                wg.Done()
                                                return
                                        }
                                }()
                                wg.Wait()
                                outdataset.Shapes = append(outdataset.Shapes, feature)
                        }

                }()
        }

        wg.Wait()

        return err
}


//parseGEOJSONAttributes ...
func parseGEOJSONAttributes(gfeature *FeatureInfo) []Attribute {
        var atts []Attribute
        for k, v := range gfeature.geojson.Properties {
                switch v {
                case nil, "", 0, "0":
                        delete(gfeature.geojson.Properties, k)
                        continue
                }
                switch k {
                case "name":
                        gfeature.name = fmt.Sprintf("%v", v)
                case "styletype":
                        gfeature.styletype = fmt.Sprintf("%v", v)
                default:
                        var attrib Attributes
                        attrib.Key = k
                        attrib.Value = fmt.Sprintf("%v", v)
                        atts = append(atts, attrib)
                }
        }
        return atts
}


//deriveGEOJSONPoints ...
func deriveGEOJSONPoints(gfeature *FeatureInfo, ch chan Point) PointArray {
        var pointarray PointArray

        switch gfeature.geojson.Geometry.Type {

        case "Point","PointZ","point","pointz":
                var z float64
                x, y := To3857(gfeature.geojson.Geometry.Point[0], gfeature.geojson.Geometry.Point[1])
                if x == 0 && y == 0 {
                        return pointarray
                }
                if len(gfeature.geojson.Geometry.Point) < 3 {
                        z, _ = getElev(x, y)
                } else {
                        z = gfeature.geojson.Geometry.Point[2]
                }

                // keep a collective of the min / max coords of dataset
                var point Coordinate
                point.X = x
                point.Y = y
                ch <- point

                pointarray.Points = append(pointarray.Points, []float64{x, y, z})

        case "LineString","LineStringZ","linestring","linestringz":
                var z float64
                for _, coords := range gfeature.geojson.Geometry.LineString {
                        x, y := To3857(coords[0], coords[1])
                        if x == 0 && y == 0 {
                                continue
                        }
                        if len(coords) < 3 {
                                z, _ = getElev(x, y)
                        } else {
                                z = coords[2]
                        }

                        // keep a collective of the min / max coords of dataset
                        var point Coordinate
                        point.X = x
                        point.Y = y
                        ch <- point

                        pointarray.Points = append(pointarray.Points, []float64{x, y, z})
                }

        case "Polygon","PolygonZ","polygon","polygonz":
                var z float64
                for _, coords := range gfeature.geojson.Geometry.Polygon {
                        for _, coord := range coords {
                                x, y := To3857(coord[0], coord[1])
                                if x == 0 && y == 0 {
                                        continue
                                }
                                if len(coord) < 3 {
                                        z, _ = getElev(x, y)
                                } else {
                                        z = coord[2]
                                }

                                // keep a collective of the min / max coords of dataset
                                var point Coordinate
                                point.X = x
                                point.Y = y
                                ch <- point

                                pointarray.Points = append(pointarray.Points, []float64{x, y, z})
                        }
                }

        }
        return pointarray
}

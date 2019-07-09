package convert

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/golang/geo/s2"
	"github.com/paulmach/go.geo"
	"github.com/paulmach/go.geojson"
	"github.com/remeh/sizedwaitgroup"
)

// Datasets ...
type Datasets struct {
	ID      string       `json:"id" yaml:"id"`
	Name    string       `json:"name" yaml:"name"`
	Url     string       `json:"dataurl" yaml:"dataurl"`
	Updated string       `json:"lastUpdated" yaml:"lastUpdated"`
	Center  []Point      `json:"center" yaml:"center"`
	S2      []string     `json:"s2" yaml:"s2"`
	Points  []Points     `json:"points" yaml:"points"`
	Lines   []Lines      `json:"lines" yaml:"lines"`
	Shapes  []Shapes     `json:"shapes" yaml:"shapes"`
}

// Individual Point Coordinate ...
type Point struct {
	X float64 `json:"x" yaml:"x"`
	Y float64 `json:"y" yaml:"y"`
	Z float64 `json:"z" yaml:"z"`
}

// Points ...
type Points struct {
	ID         string      `json:"id" yaml:"id"`
	Name       string      `json:"name" yaml:"name"`
	StyleType  string      `json:"type" yaml:"type"`
	Attributes []Attribute `json:"attributes" yaml:"attributes"`
	Points     []float64   `json:"point" yaml:"point"`
}

// PointArrays ...
type PointArray struct {
	Points     [][]float64 `json:"points" yaml:"points"`
}

// Line ...
type Lines struct {
	ID         string      `json:"id" yaml:"id"`
	Name       string      `json:"name" yaml:"name"`
	StyleType  string      `json:"type" yaml:"type"`
	Attributes []Attribute `json:"attributes" yaml:"attributes"`
	Points     [][]float64 `json:"points" yaml:"points"`
}

// Shape ...
type Shapes struct {
	ID         string      `json:"id" yaml:"id"`
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

// FeatureInfo ...
type FeatureInfo struct {
	ID        string
	Geojson   geojson.Feature
	GeomType  string
	SRID      string
	S2        []s2.CellID
	Tokens    []string
	Name      string
	StyleType string
}

// BBOX ExtentContainer
type ExtentContainer struct {
	bbox	map[string]float64
	ch	chan []float64
	wg	sizedwaitgroup.SizedWaitGroup
}

const (
	// env var for the dem.ver path
	envDEMVRT = "DEMVRT"

	// process limits for the sized wait group
	maxRoutines = 50
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
		dvp = path.Join(cwd, "earthdem.vrt")
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

	//store the csv headers by index
	headers := make(map[int]string)
        container := initExtentContainer()

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
			container.wg.Add()
			go ParseCSV(headers, record, &outdataset, container)
		}
	}

	container.wg.Wait()

	// close the BBOXlistener goroutine
	close(container.ch)

	// configure the center point... in 4326
	c := getCenter(container.bbox)
	outdataset.Center = append(outdataset.Center, c)

	// configure the s2 array... in 4326
	outdataset.S2 = s2covering(container.bbox)

	return &outdataset, err
}

// DatasetFromGEOJSON ...
func DatasetFromGEOJSON(xField string, yField string, zField string, contents io.Reader) (*Datasets, error) {
	raw, err := ioutil.ReadAll(contents)
	if err != nil {
		return nil, err
	}

	//carries references to this dataset's ch, wg, and bbox
        container := initExtentContainer()

	rawjson, err := geojson.UnmarshalFeatureCollection(raw)

	// this kicks off the processing of the data
	outdataset, err := parseGEOJSONCollection(rawjson, container)
	if err != nil {
		return outdataset, err
	}

	// close the BBOXlistener goroutine
        close(container.ch)

	// configure the center point... in 4326
	c := getCenter(container.bbox)
	outdataset.Center = append(outdataset.Center, c)

	// configure the s2 array... in 4326
	outdataset.S2 = s2covering(container.bbox)

	return outdataset, err
}

// ParseCSV ...
func ParseCSV(headers map[int]string, record []string, outdataset *Datasets, container *ExtentContainer) {

	defer container.wg.Done()

	var xyz []float64
	var point Points

	for i, value := range record {
		switch headers[i] {
		case "X":
			x, _ := strconv.ParseFloat(value, 64)
			xyz = append(xyz, x)
		case "Y":
			y, _ := strconv.ParseFloat(value, 64)
			xyz = append(xyz, y)
		case "Z":
			z, _ := strconv.ParseFloat(value, 64)
			xyz = append(xyz, z)
		default:
			var atts Attribute
			atts.Key = headers[i]
			atts.Value = fmt.Sprintf("%v", value)
			point.Attributes = append(point.Attributes, atts)
		}
	}

	// enforce 3857 and elevation
	coord := checkCoords(xyz)

	// keep a collective of the min / max coords of dataset
	container.ch <- coord

	// fill in the poiiint float array
	point.Points = append(point.Points, coord[0], coord[1], coord[2])

	// finally, append point to the final dataset
	outdataset.Points = append(outdataset.Points, point)
}

// BBOXListener ...  observes every X & Y on the channel, retains lowest and highest for bbox extent
func BBOXListener(container *ExtentContainer) {

	for {
		xyz, ok := <-container.ch

		// if channel closes, kill goroutine
		if !ok {
			return
		}

                X := xyz[0]
                Y := xyz[1]

		_, present := container.bbox["lx"]
		if !present {
			container.bbox["lx"] = X
			container.bbox["rx"] = X
			container.bbox["ly"] = Y
			container.bbox["uy"] = Y
		}

		// if the inbound X is outside of current extent, grow extent
		if X < container.bbox["lx"] {
			container.bbox["lx"] = X
		} else if X > container.bbox["rx"] {
			container.bbox["ux"] = X
		}

		// if the inbound Y is outside of current extent, grow extent
		if Y < container.bbox["ly"] {
			container.bbox["ly"] = Y
		} else if Y > container.bbox["uy"] {
			container.bbox["uy"] = Y
		}
	}
}

// getCenter calculates the center of a bbox extent
func getCenter(bbox map[string]float64) Point {
	var c Point
        c.X = bbox["rx"] - (bbox["rx"]-bbox["lx"])/2
        c.Y = bbox["uy"] - (bbox["uy"]-bbox["ly"])/2
        c.Z, _ = GetElev(c.X, c.Y)

	return c
}

// s2covering finds the s2 hash key that represents the geographic coverage of the bbox extent
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

		// write s2 token array to the dataset s2 key list
		s2hash = append(s2hash, token)
	}

	return s2hash
}

// GetElev gets the elevation for the given x y coordinate
func GetElev(x float64, y float64) (float64, error) {
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

// To4326 converts coordinates to EPSG:4326 projection
func To4326(x float64, y float64) (float64, float64) {
	if x >= 180 || x <= -180 {
		mercPoint := geo.NewPoint(x, y)
		geo.Mercator.Inverse(mercPoint)
		x = mercPoint[0]
		y = mercPoint[1]
	}

	return x, y
}

// To3857 converts coordinates to EPSG:3857 projection
func To3857(x float64, y float64) (float64, float64) {
	if x <= 180 || x >= -180 {
		mercPoint := geo.NewPoint(x, y)
		geo.Mercator.Project(mercPoint)
		x = mercPoint[0]
		y = mercPoint[1]

		// trim decimals to the cm
		x = math.Round(mercPoint[0]*100) / 100
		y = math.Round(mercPoint[1]*100) / 100
	}

	return x, y
}

//ParseGEOJSONCollection peels into the collection multiple features
func parseGEOJSONCollection(collection *geojson.FeatureCollection, container *ExtentContainer) (*Datasets, error) {
	var outdataset Datasets
	var err error

	if len(collection.Features) < 1 {
		return &outdataset, err
	}

	//access each of the individual features of the geojson
	for _, item := range collection.Features {

		// the new feature
		var gfeature FeatureInfo
		gfeature.Geojson = *item

		container.wg.Add()

		// set off a new go routine for each feature
		go func() {
			defer container.wg.Done()

			// process each feature independently
			ParseGEOJSONFeature(&gfeature, &outdataset, container)
		}()
	}

	container.wg.Wait()

	return &outdataset, err
}

//ParseGEOJSONFeature processes each geojson feature into a Unity json feature
func ParseGEOJSONFeature (gfeature *FeatureInfo, outdataset *Datasets, container *ExtentContainer) {
	log.Printf("parsing geojson feature....")
        switch gfeature.Geojson.Geometry.Type {

                // it appears the following is replicate, but with type asserstion and
                // minute differences, the least complicated path is to replicate some
                // elements.
                case "Point", "Pointz","POINT":
                        var wg sync.WaitGroup
                        var feature Points
                        wg.Add(2)
                        go func() {
                                defer wg.Done()
                                feature.Attributes = ParseGEOJSONAttributes(gfeature)
                                feature.Name = gfeature.Name
                                feature.StyleType = gfeature.StyleType
                                feature.ID = gfeature.ID
                        }()
			go func() {
                                defer wg.Done()
                                feature.Points = ParseGEOJSONGeom(gfeature,container).Points[0]
                                if len(feature.Points) < 1 {
                                        return
                                }
                        }()
                        //feature.Points = (gfeature.Geojson.Geometry.Point)
                        wg.Wait()
                        outdataset.Points = append(outdataset.Points, feature)

                case "LineString","LineStringZ","LINESTRING":
                        var wg sync.WaitGroup
                        var feature Lines
                        wg.Add(2)
                        go func() {
                                defer wg.Done()
                                feature.Attributes = ParseGEOJSONAttributes(gfeature)
                                feature.Name = gfeature.Name
                                feature.StyleType = gfeature.StyleType
                                feature.ID = gfeature.ID
                        }()
                        go func() {
				defer wg.Done()
				feature.Points = ParseGEOJSONGeom(gfeature,container).Points
				if len(feature.Points) < 1 {
					return
				}
			}()
                        //feature.Points = gfeature.Geojson.Geometry.LineString
                        wg.Wait()
                        outdataset.Lines = append(outdataset.Lines, feature)

                case "Polygon","PolygonZ","POLYGON":
                        var wg sync.WaitGroup
                        var feature Shapes
                        wg.Add(2)
                        go func() {
                                defer wg.Done()
                                feature.Attributes = ParseGEOJSONAttributes(gfeature)
                                feature.Name = gfeature.Name
                                feature.StyleType = gfeature.StyleType
                                feature.ID = gfeature.ID
                        }()
			go func() {
				defer wg.Done()
				feature.Points = ParseGEOJSONGeom(gfeature,container).Points
				if len(feature.Points) < 1 {
					return
				}
			}()
                        //feature.Points = gfeature.Geojson.Geometry.Polygon[0]
                        wg.Wait()
                        outdataset.Shapes = append(outdataset.Shapes, feature)

                }
}

// ParseGEOJSONAttributes cleans & prepares all attributes
func ParseGEOJSONAttributes(gfeature *FeatureInfo) []Attribute {
        var atts []Attribute
        for k, v := range gfeature.Geojson.Properties {

                // by using switch on v, we don't need to reflect the interface.TypeOf()
                switch v {
                        case nil, "", 0, "0":
                                delete(gfeature.Geojson.Properties, k)
                                continue
                }

                // for the remaining keys with values....
                switch k {

                        // v requires type assertion
                        case "name":
                                gfeature.Name = fmt.Sprintf("%v",v)
                        case "styletype":
                                gfeature.StyleType = fmt.Sprintf("%v",v)
                        case "id","fid","osm_id","uid","uuid":
                                gfeature.ID = fmt.Sprintf("%v",v)
                        default:
                                var attrib Attribute
                                attrib.Key = k
                                attrib.Value = fmt.Sprintf("%v",v)
                                atts = append(atts, attrib)
                }
        }
        return atts
}

//ParseGEOJSONGeom cleans & prepares the geometry, filling in Z values if absent
func ParseGEOJSONGeom(gfeature *FeatureInfo, container *ExtentContainer) PointArray {
	var pointarray PointArray

	// subsequently complex geometry types require traversing nested geometries
	switch gfeature.Geojson.Geometry.Type {

	case "Point", "Pointz", "POINT":
		point := checkCoords(gfeature.Geojson.Geometry.Point)
		if container != nil {container.ch <- point}
                pointarray.Points = append(pointarray.Points, point)

	case "LineString", "LineStringz","LINESTRING":
		for _, coord := range gfeature.Geojson.Geometry.LineString {
			point := checkCoords(coord)
			if container != nil {container.ch <- point}
                        pointarray.Points = append(pointarray.Points, point)
		}

	case "Polygon", "Polygonz","POLYGON":
		for _, coords := range gfeature.Geojson.Geometry.Polygon {
			for _, coord := range coords {
				point := checkCoords(coord)
				if container != nil {container.ch <- point}
				pointarray.Points = append(pointarray.Points, point)
			}
		}

	}

	return pointarray
}

// checkCoords ... enforces 3857 for X and Y, and fills Z if absent
func checkCoords (coord []float64) []float64 {

	// ommit coords that are malformed (no x and y, or more than xyz)
	if len(coord) == 0 {
		return coord
	} else if len(coord) > 2  && coord[2] != 0 {
		return coord
	}

	var z float64

	x, y := To3857(coord[0], coord[1])

	// check for z value
	if len(coord) < 3 {
		z, _ = GetElev(x, y)
	} else {
		z = coord[2]
	}

	return []float64{x, y, z}
}

// initExtentContainer sets up all the elements of the empty struct
func initExtentContainer () *ExtentContainer {
        var container ExtentContainer

        // the bbox extent that will observe and grow with coordinates
        container.bbox = make(map[string]float64)

        // the channel that carries the coordinates synchronously
        container.ch = make(chan []float64)

        // a wait group if sub go routines need to add to total
        wg := sizedwaitgroup.New(maxRoutines)
        container.wg = wg

        // the bbox extent listener on the channel doing work with the coords
        go BBOXListener(&container)

        return &container
}


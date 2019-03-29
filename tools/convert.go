package converter

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	shapes "github.com/jonas-p/go-shp"
	"github.com/golang/geo/s2"
	geo "github.com/paulmach/go.geo"
	geojson "github.com/paulmach/go.geojson"
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
	geojson   geojson.Feature //G
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
func DatasetFromCSV(xField string, yField string, zField string, contents []byte) (*Datasets, error) {
	raw, err := csv.NewReader(bytes.NewReader(contents)).ReadAll()
	if err != nil {
		return nil, err
	}

	var outdataset Datasets

	headers := make(map[int]string)
	bbox := make(map[string]float64)

	for i, record := range raw {
		var pointxyz Coordinate
		var point Point
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

			for i, value := range record {
				switch headers[i] {
				case "X":
					pointxyz.X, _ = strconv.ParseFloat(value, 64)
				case "Y":
					pointxyz.Y, _ = strconv.ParseFloat(value, 64)
				case "Z":
					pointxyz.Z, _ = strconv.ParseFloat(value, 64)

					MinMax(bbox, pointxyz.X, pointxyz.Y)

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
			}

			//finally, fill in the point float array
			point.Point = append(point.Point, pointxyz.X, pointxyz.Y, pointxyz.Z)
		}

		outdataset.Points = append(outdataset.Points, point)
	}

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

// DatasetFromSHP.....
func DatasetFromSHP(shapeFile []byte, dbfFile []byte) (*Datasets, error) {
	// a 'shp' file is actually a zip of a shp and a dbf
	// shapefiles use fieldnames of X Y Z for the geom, no add'l fields required

	raw := &shapes.Reader{shp: shapeFile, dbf: dbfFile}
	raw.readHeaders()

	// abridged function to read dbf w/out passing filename
	// read header
	raw.dbf.Seek(4, io.SeekStart)
	binary.Read(raw.dbf, binary.LittleEndian, &raw.dbfNumRecords)
	binary.Read(raw.dbf, binary.LittleEndian, &raw.dbfHeaderLength)
	binary.Read(raw.dbf, binary.LittleEndian, &raw.dbfRecordLength)

	raw.dbf.Seek(20, io.SeekCurrent) // skip padding
	numFields := int(math.Floor(float64(raw.dbfHeaderLength-33) / 32.0))
	raw.dbfFields = make([]Field, numFields)
	binary.Read(raw.dbf, binary.LittleEndian, &raw.dbfFields)

	var outdataset Datasets

	for shape.Next() {
		n, p := shape.Shape()
		// print feature
		fmt.Println(reflect.TypeOf(p).Elem(), p.BBox())

		// print attributes
		for k, f := range fields {
			val := shape.ReadAttribute(n, k)
			fmt.Printf("\t%v: %v\n", f, val)
		}
	}

	return

	headers := make(map[int]string)
	bbox := make(map[string]float64)

	for i, record := range raw {
		var pointxyz Coordinate
		var point Point
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

			for i, value := range record {
				switch headers[i] {
				case "X":
					pointxyz.X, _ = strconv.ParseFloat(value, 64)
				case "Y":
					pointxyz.Y, _ = strconv.ParseFloat(value, 64)
				case "Z":
					pointxyz.Z, _ = strconv.ParseFloat(value, 64)

					MinMax(bbox, pointxyz.X, pointxyz.Y)

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
			}

			//finally, fill in the point float array
			point.Point = append(point.Point, pointxyz.X, pointxyz.Y, pointxyz.Z)
		}

		outdataset.Points = append(outdataset.Points, point)
	}

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

// MinMax ...
func MinMax(bbox map[string]float64, X float64, Y float64) {
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
	if (x > 180) || (x < -180) {
		return x, y
	}

	mercPoint := geo.NewPoint(x, y)
	geo.Mercator.Project(mercPoint)
	return mercPoint[0], mercPoint[1]
}

package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"github.com/golang/geo/s2"
	"gopkg.in/yaml.v2"
	//    "github.com/jonas-p/go-shp"
)

func dataHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	_, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	check(err)

	reader := multipart.NewReader(r.Body, params["boundary"])

	var indataset Input
	var converted []byte
	var data []byte
	var info []byte

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}

		switch part.FormName() {
		case "info":
			info, err = ioutil.ReadAll(part)
			check(err)
			fmt.Printf("%s\n", info)
			fmt.Printf("read from info!\n")
		case "file":
			data, err = ioutil.ReadAll(part)
			check(err)
			fmt.Printf("read from data!\n")
		}
	}

	err = yaml.Unmarshal(info, &indataset)
	check(err)

	format := indataset.Format

	switch format {
	case "csv":
		converted, err = CsvHandler(indataset, data)
		check(err)
	//case "shp":
	//outdataset, err = ShpHandler(indataset, data)
	//case "dxf":
	//outdataset, err = DxfHandler(indataset, contents)
	default:
		converted = []byte("Sorry, things didn't work out.  Is the format supported?")
	}

	//    ioutil.WriteFile("tests/out.json", converted, 0644)
	w.Write(converted)

	log.Println("total dataset round trip:", int64(time.Since(start).Seconds()*1e3), "ms")
}

// trying to convert a csv?  here's where it happens
func CsvHandler(indataset Input, contents []byte) (converted []byte, err error) {
	log.Println("request for a csv conversion")

	start := time.Now()
	s := bytes.NewReader(contents)

	raw, err := csv.NewReader(s).ReadAll()
	check(err)

	xfield := indataset.Xfield
	yfield := indataset.Yfield
	zfield := indataset.Zfield

	var outdataset Datasets
	headers := make(map[int]string)
	bbox := make(map[string]float64)

	for i, record := range raw {
		var pointxyz Point
		var point Points
		switch i {
		case 0:
			for i, header := range record {
				switch header {
				case xfield:
					headers[i] = "X"
				case yfield:
					headers[i] = "Y"
				case zfield:
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
					var atts Attributes
					atts.Key = headers[i]
					atts.Value = fmt.Sprintf("%v", value)
					//TBD geojson pair := make(map[string]interface{})
					//TBD geojson pair[headers[i]] = value
					point.Attributes = append(point.Attributes, atts)
				}
			}

			// fill elevation for the processing node of the point, line, or shape if required
			if pointxyz.Z == 0 && pointxyz.X != 0 && pointxyz.Y != 0 {
				log.Printf("value needed filling in with elevation...")
				pointxyz.Z, err = getElev(pointxyz.X, pointxyz.Y)
			}

			//finally, fill in the point float array
			point.Point = append(point.Point, pointxyz.X, pointxyz.Y, pointxyz.Z)
		}

		outdataset.Points = append(outdataset.Points, point)
	}

	// configure the center point... in 4326
	var c Point
	c.X = bbox["rx"] - (bbox["rx"]-bbox["lx"])/2
	c.Y = bbox["uy"] - (bbox["uy"]-bbox["ly"])/2
	c.Z, _ = getElev(c.X, c.Y)
	outdataset.Center = append(outdataset.Center, c)

	// configure the s2 array... in 4326
	outdataset.S2 = s2covering(bbox)

        // finally, process into the unity json struct
	converted, err = json.Marshal(outdataset)

        // add that we've processed a new csv dataset to the counter
	counter.Incr("csv")
	log.Println("csv's processed:", counter.Get("csv"), ", time:", int64(time.Since(start).Seconds()*1e3), "ms")
	return converted, err
}

// any time a dataset comes in.... the output unity json REQUIRES a minmax in lat long (for tap to zoom) 
func MinMax(bbox map[string]float64, X float64, Y float64) {
	_, ok := bbox["lx"]
	if !ok {
		bbox["lx"] = X
		bbox["rx"] = X
		bbox["ly"] = Y
		bbox["uy"] = Y
	}

	switch {
	case X < bbox["lx"]:
		bbox["lx"] = X
	case X > bbox["rx"]:
		bbox["ux"] = X
	}

	switch {
	case Y < bbox["ly"]:
		bbox["ly"] = Y
	case Y > bbox["uy"]:
		bbox["uy"] = Y
	}
}

// any time a dataset comes in.... the output unity json is set to use s2 covering (for lots of reasons too many to discuss here)
// s2 coverings are badass, new google tech, and what the nearme service also relies upon
func s2covering(bbox map[string]float64) []string {
	var s2hash []string

	rx, uy := To4326(bbox["rx"], bbox["uy"])
	lx, ly := To4326(bbox["lx"], bbox["ly"])
	cz, err := getElev(bbox["rx"], bbox["uy"])
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(lx, ly, rx, uy)

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
		if tokencheck(token, s2hash) == false {
			fmt.Printf(token)
			s2hash = append(s2hash, token)
		}
	}

	return s2hash
}

// token checking is not something that belongs here at all.   can remove
func tokencheck(token string, list []string) bool {
	for _, b := range list {
		if b == token {
			return true
		}
	}
	return false
}

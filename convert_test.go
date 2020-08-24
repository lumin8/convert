package convert

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

const (
	//csv testing datasets
	pointswithZ       = "tests/trek/trek_drilldata.csv"
	pointswithZ_input = "tests/trek/trek_drilldata.json"

	pointsnoZ       = "tests/bonanza/bonanza_soils.csv"
	pointsnoZ_input = "tests/bonanza/bonanza_soils.json"

	points4326       = "tests/bonanza/bonanza_soils_4326.csv"
	points4326_input = "tests/bonanza/bonanza_soils_4326.json"

	fakepoints       = "tests/fake/fake_soils.csv"
	fakepoints_input = "tests/fake/fake_soils.json"

	//geojson testing datasets
	pointsgeojson       = "tests/bonanza/bonanza_soils.geojson"
	pointsgeojson_input = "tests/bonanza/bonanza_lines.json"

	fakecoords       = "tests/fake/fake_coords.geojson"
	fakecoords_input = "tests/fake/fake_coords.geojson"

	lines       = "tests/bonanza/bonanza_lines.geojson"
	lines_input = "tests/bonanza/bonanza_lines.json"

	largeshapes       = "tests/bonanza/bonanza_formations.geojson"
	largeshapes_input = "tests/bonanza/bonanza_formations.json"

	manyshapes       = "tests/bonanza/bonanza_outcrops.geojson"
	manyshapes_input = "tests/bonanza/bonanza_outcrops.json"

	singleshape       = "tests/fake/testshape.geojson"
        singleshape_input = "tests/fake/testshape.geojson"

        singleshape3D       = "tests/fake/testshape3D.geojson"
        singleshape3D_input = "tests/fake/testshape3D.geojson"

        singlemultielev       = "tests/bonanza/bonanza_multiwithelev.geojson"
        singlemultielev_input = "tests/bonanza/bonanza_multiwithelev.json"

	//kml testing datasets
	pointskml = "tests/kml/points.kml"
	lineskml  = "tests/kml/lines.kml"
	shapeskml = "tests/kml/shapes.kml"

	//gpx testing datasets
	points3Dgpx = "tests/gpx/points3D.gpx"
	linesgpx    = "tests/gpx/lines.gpx"
	shapesgpx   = "tests/gpx/lines3D.gpx"
)

type Input struct {
	Xfield string `json:"xfield" yaml:"xfield"`
	Yfield string `json:"yfield" yaml:"yfield"`
	Zfield string `json:"zfield" yaml:"zfield"`
	Srs    int    `json:"srs" yaml:"srs"`
}

func TestCSVData(t *testing.T) {

	// build a map of the testing data and inputs
	data := make(map[string]string)
	data[pointswithZ] = pointswithZ_input
	data[pointsnoZ] = pointsnoZ_input
	data[points4326] = points4326_input
	data[fakepoints] = fakepoints_input

	for item, inputDetails := range data {

		// prase the inputDetails from originating json
		jsonFile, err := os.Open(inputDetails)
		if err != nil {
			t.Errorf(err.Error())
		}
		byteValue, _ := ioutil.ReadAll(jsonFile)
		var input Input
		json.Unmarshal(byteValue, &input)

		// grab the item as a reader
		data, err := os.Open(item)
		if err != nil {
			t.Errorf(err.Error())
		}

		// send the information to the tester
		results, err := DatasetFromCSV(input.Xfield, input.Yfield, input.Zfield, data)
		if err != nil {
			t.Logf("csv conversion error for %s, no features in dataset: %s\n", item, err.Error())
		}

		// parse the results
		final, err := json.Marshal(results)
		if err != nil {
			t.Errorf("jsor marshal error for %s: %s\n", item, err.Error())
		}

		// guessing that the final string should be more than 100 characters
		if results == nil {
			t.Logf("no valid features were found for %s:%v\n", item, final)
			return
		}

		fmt.Printf("conversion for %s was successful, result center is %v\n", item, results.Center)

	}
}

func TestGEOJSONData(t *testing.T) {

	// build a map of the testing data and inputs
	data := make(map[string]string)
	data[pointsgeojson] = pointsgeojson_input
	data[lines] = lines_input
	data[largeshapes] = largeshapes_input         //polygons
	data[manyshapes] = manyshapes_input           //multipolygons
	data[singleshape] = singleshape_input         //a single very large polygon
	data[singlemultielev] = singlemultielev_input // multi nearly vertical

	for item, inputDetails := range data {

		// prase the inputDetails from originating json
		jsonFile, err := os.Open(inputDetails)
		if err != nil {
			t.Errorf(err.Error())
		}
		byteValue, _ := ioutil.ReadAll(jsonFile)
		var input Input
		json.Unmarshal(byteValue, &input)

		// grab the item as a reader
		data, err := os.Open(item)
		if err != nil {
			t.Errorf(err.Error())
		}

		// send the information to the tester
		results, err := DatasetFromGEOJSON(input.Xfield, input.Yfield, input.Zfield, data)

		if err != nil {
			t.Errorf("[DatasetFromGEOJSON] in pkg [convert], geojson conversion error for %s: %s\n", item, err.Error())
		}

		// parse the results
		final, err := json.Marshal(results)
		if err != nil {
			t.Errorf("json marshal error for %s: %s\n", item, err.Error())
		}

		// if no center, the conversion is BUNK
		// guessing that the final string should be more than 100 characters
		if results == nil {
			t.Logf("no valid features were found for %s:%v\n", item, final)
			return
		}

		fmt.Printf("conversion for %s was successful, result center is %v\n", item, results.Center)

		// the following prints out the file product, useful for debugging only
		err = ioutil.WriteFile(inputDetails+".outfile", final, 0644)
		if err != nil {
			t.Errorf(err.Error())
		}

	}
}

func TestKMLData(t *testing.T) {

	// build a map of the testing data and inputs
	data := make(map[string]string)
	data[pointskml] = pointskml
	data[lineskml] = lineskml
	data[shapeskml] = shapeskml

	for item, inputDetails := range data {

		// grab the item as a reader
		data, err := os.Open(item)
		if err != nil {
			t.Errorf(err.Error())
		}

		// send the information to the tester
		results, err := DatasetFromKML("", "", "", data)

		if err != nil {
			t.Errorf("[DatasetFromKML] in pkg [convert], kml conversion error for %s: %s\n", item, err.Error())
		}

		// parse the results
		final, err := json.Marshal(results)
		if err != nil {
			t.Errorf("json marshal error for %s: %s\n", item, err.Error())
		}

		// if no center, the conversion is BUNK
		// guessing that the final string should be more than 100 characters
		if results == nil {
			t.Logf("no valid features were found for %s:%v\n", item, final)
			return
		}

		fmt.Printf("conversion for %s was successful, result center is %v\n", item, results.Center)

		// the following prints out the file product, useful for debugging only
		err = ioutil.WriteFile(inputDetails+".outfile", final, 0644)
		if err != nil {
			t.Errorf(err.Error())
		}

	}
}

func TestGPXData(t *testing.T) {

	// build a map of the testing data and inputs
	data := make(map[string]string)
	data[points3Dgpx] = points3Dgpx
	data[linesgpx] = linesgpx
	data[shapesgpx] = shapesgpx

	for item, inputDetails := range data {

		// grab the item as a reader
		data, err := os.Open(item)
		if err != nil {
			t.Errorf(err.Error())
		}

		// send the information to the tester
		results, err := DatasetFromGPX("", "", "", data)

		if err != nil {
			t.Errorf("[DatasetFromGPX] in pkg [convert], gpx conversion error for %s: %s\n", item, err.Error())
		}

		// parse the results
		final, err := json.Marshal(results)
		if err != nil {
			t.Errorf("json marshal error for %s: %s\n", item, err.Error())
		}

		// if no center, the conversion is BUNK
		// guessing that the final string should be more than 100 characters
		if results == nil {
			t.Logf("no valid features were found for %s:%v\n", item, final)
			return
		}

		fmt.Printf("conversion for %s was successful, result center is %v\n", item, results.Center)

		// the following prints out the file product, useful for debugging only
		err = ioutil.WriteFile(inputDetails+".outfile", final, 0644)
		if err != nil {
			t.Errorf(err.Error())
		}

	}
}

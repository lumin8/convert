package convert

import (
	"testing"
	"fmt"
	"io/ioutil"
	"os"
	"encoding/json"
)

const (
	//csv testing datasets
	pointswithZ = "tests/trek/trek_drilldata.csv"
	pointswithZ_input = "tests/trek/trek_drilldata.json"

	pointsnoZ = "tests/bonanza/bonanza_soils.csv"
	pointsnoZ_input = "tests/bonanza/bonanza_soils.json"

	points4326 = "tests/bonanza/bonanza_soils_4326.csv"
        points4326_input = "tests/bonanza/bonanza_soils_4326.json"

	fakepoints = "tests/fake/fake_soils.csv"
	fakepoints_input = "tests/fake/fake_soils.json"

	//geojson testing datasets
	pointsgeojson = "tests/bonanza/bonanza_soils.geojson"
	pointsgeojson_input = "tests/bonanza/bonanza_lines.json"

	fakecoords = "tests/fake/fake_coords.geojson"
        fakecoords_input = "tests/fake/fake_coords.geojson"

	lines = "tests/bonanza/bonanza_lines.geojson"
	lines_input = "tests/bonanza/bonanza_lines.json"

	shapes = "tests/bonanza/bonanza_formations.geojson"
	shapes_input = "tests/bonanza/bonanza_formations.json"
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
			t.Logf(err.Error())
			t.Fail()
		}
		byteValue, _ := ioutil.ReadAll(jsonFile)
		var input Input
		json.Unmarshal(byteValue, &input)


		// grab the item as a reader
		data, err := os.Open(item)
		if err != nil {
                        t.Logf(err.Error())
                        t.Fail()
                }

		// send the information to the tester
		results, err := DatasetFromCSV(input.Xfield, input.Yfield, input.Zfield, data)
		if err != nil {
			t.Logf("csv conversion error for %s, no features in dataset: %s\n",item,err.Error())
			t.Fail()
		}

		// parse the results
		final, err := json.Marshal(results)
		if err != nil {
                        t.Logf("json marshal error for %s: %s\n",item,err.Error())
			t.Fail()
                }

		// guessing that the final string should be more than 100 characters
		if results == nil {
			t.Logf("no valid features were found for %s:%v\n",item,final)
			return
		}

		fmt.Printf("conversion for %s was successful, with a center of %v\n",item,results.Center)

	}
}


func TestGEOJSONData(t *testing.T) {

        // build a map of the testing data and inputs
        data := make(map[string]string)
	data[pointsgeojson] = pointsgeojson_input
	data[fakecoords] = fakecoords_input
        data[lines] = lines_input
        data[shapes] = shapes_input

        for item, inputDetails := range data {

                // prase the inputDetails from originating json
                jsonFile, err := os.Open(inputDetails)
                if err != nil {
                        t.Logf(err.Error())
                        t.Fail()
                }
                byteValue, _ := ioutil.ReadAll(jsonFile)
                var input Input
                json.Unmarshal(byteValue, &input)

                // grab the item as a reader
		// grab the item as a reader
                data, err := os.Open(item)
                if err != nil {
                        t.Logf(err.Error())
                        t.Fail()
                }

                // send the information to the tester
                results, err := DatasetFromGEOJSON(input.Xfield, input.Yfield, input.Zfield, data)

                if err != nil {
                        t.Logf("geojson conversion error, no valid features for %s: %s\n",item,err.Error())
			t.Fail()
                }

		// parse the results
		final, err := json.Marshal(results)
                if err != nil {
                        t.Logf("json marshal error for %s: %s\n",item,err.Error())
                        t.Fail()
                }

                // if no center, the conversion is BUNK
		// guessing that the final string should be more than 100 characters
                if results == nil {
                        t.Logf("no valid features were found for %s:%v\n",item,final)
                        return
                }

                fmt.Printf("conversion for %s was successful, result center is %v\n",item,results.Center)
        }
}

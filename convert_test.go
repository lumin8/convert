package convert

import (
	"testing"
	"fmt"
	"io/ioutil"
	"os"
	"encoding/json"
)

const (
	pointswithZ = "tests/trek/trek_drilldata.csv"
	pointswithZ_input = "tests/trek/trek_drilldata.json"

	pointsnoZ = "tests/bonanza/bonanza_soils.csv"
	pointsnoZ_input = "tests/bonanza/bonanza_soils.json"

	//pointsgeojson = "tests/bonanza/bonanza_points.geojson"
	//pointsgeojson_input = "tests/bonanza/bonanza_points.json"

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

	for item, inputDetails := range data {

		// prase the inputDetails from originating json
		jsonFile, err := os.Open(inputDetails)
		if err != nil {
			fmt.Println(err)
			continue
		}
		byteValue, _ := ioutil.ReadAll(jsonFile)
		var input Input
		json.Unmarshal(byteValue, &input)


		// grab the item as a reader
		data, err := os.Open(item)
		if err != nil {
                        fmt.Println(err)
                        continue
                }

		// send the information to the tester
		results, err := DatasetFromCSV(input.Xfield, input.Yfield, input.Zfield, data)
		if err != nil {
			fmt.Printf("damnit, got an error : %s\n",err.Error())
			continue
		}

		// parse the results
		final, err := json.Marshal(results)
		if err != nil {
                        fmt.Printf("damnit, got an error : %s\n",err.Error())
                        continue
                }

		// guessing that the final string should be more than 100 characters
		if len(final) < 100 {
			fmt.Printf("damnit, conversion didn't work: %s\n",err.Error())
                        continue
		}

		fmt.Printf("conversion for %s was successful\n",item)
		//fmt.Printf(string(final))

	}
}


func TestGEOJSONData(t *testing.T) {

        // build a map of the testing data and inputs
        data := make(map[string]string)
        data[lines] = lines_input
        data[shapes] = shapes_input

        for item, inputDetails := range data {

                // prase the inputDetails from originating json
                jsonFile, err := os.Open(inputDetails)
                if err != nil {
                        fmt.Println(err)
                        continue
                }
                byteValue, _ := ioutil.ReadAll(jsonFile)
                var input Input
                json.Unmarshal(byteValue, &input)

                // grab the item as a reader
		// grab the item as a reader
                data, err := os.Open(item)
                if err != nil {
                        fmt.Println(err)
                        continue
                }

                // send the information to the tester
                results, err := DatasetFromGEOJSON(input.Xfield, input.Yfield, input.Zfield, data)
                if err != nil {
                        fmt.Printf("damnit, got an error : %s\n",err.Error())
                }

		// parse the results
		final, err := json.Marshal(results)
                if err != nil {
                        fmt.Printf("damnit, got an error : %s\n",err.Error())
                        continue
                }

                // guessing that the final string should be more than 100 characters
                if len(final) < 100 {
                        fmt.Printf("damnit, conversion didn't work: %s\n",err.Error())
                        continue
                }

                fmt.Printf("conversion for %s was successful\n",item)
		//fmt.Printf(string(final))
        }
}

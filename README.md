#Data API

This api takes http POST of csv (tbd shp and dxf), and returns json or a struct used to build Unity objects for the Deep AR project.

There are currently two principle pieces: main.go and mappings.go, which hold the business end and the struct mappings end of the code project, respectively.

To use:
- download this repo
- ensure port 8000 is open (or change the config of the port in the const of main.go)
- set your GOPATH  `export GOPATH=$HOME/go`
- > go run main.go mappings.go  //or
- > ./make.bash  //compiles the program so it may be run simply by typing ./main

##Current Structure of INBOUND Data
-Multipart File

####"info" : []byte of info array
https://github.com/lumin8/deepar-data/blob/master/config/input.json

####"file" : []byte of file


##Current Structure of OUTBOUND Data
-Singlepart File

####Json String
https://github.com/lumin8/deepar-data/blob/master/config/output.json


##Tests
One test currently exists as 'data_test.go'.  This can be run from any machine, it bundles up a CSV from tests/trek/ and the json array of info (above), hits the endpoint, and receives the data back.  Currently, a 200KB csv and a nominal json expand **3x** in the current arrangement for datasets (config/output.json)... this is not ideal, further testing may indicate a need to compress the verbosity of the output.json.

## TBD
Add SHP conversion functionality
Add DSF conversion functionality
Handle different Coordinate Systems
ADd DEM Point Cloud capturing

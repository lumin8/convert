# Data API

This api takes http POST of csv (tbd shp and dxf), and returns json or a struct used to build Unity objects for the Deep AR project.

There are currently two principle pieces: main.go and mappings.go, which hold the business end and the struct mappings end of the code project, respectively.

**Current Server Address: http://map.life:8000/**

To use:
- download this repo
- ensure port 8000 is open (or change the config of the port in the const of main.go)
- set your GOPATH  `export GOPATH=$HOME/go`
- > go run main.go mappings.go  //or
- > ./make.bash  //compiles the program so it may be run simply by typing ./main


## DEM endpoint:  .../dem/?x=&y=

Hit this endpoint with a lat (y), long(x) [use negatives in the western hemisphere!], and a format (json), and what comes back will be a DEM in EPSG:3857 approximate 0.06 degrees square surrounding the central xy point.  

DEM is currently set to a 0.03 x 0.03 decimal degree wide grid, which is approximately 3km.


## NEARME endpoint:  serveraddress/nearme/?x=&y=&f=&type=

Hit this endpoint with a lat (y), long(x) [use negatives in the western hemisphere!], a format (json), and a type (see below)...  what comes back will be a dataset in EPSG:3857 approximately 2 kilometers around you (depending on the type)i...  NOTE:  all points forming these shapes WILL HAVE ELEVATION AS WELL!

types currently supported/planned:
- **&type=poi** supported=YES (points of interest, eg. peak names, waterbodies, trailheads, etc.)
- **&type=trails** supported=YES (trails, paths, hiking routes, etc.)
- **&type=roads** supported=YES (roads, interstates, etc...   this could get intense)

- **shapes** supported=SOON (building footprints, other random polygons)
- **water** supported=SOON (rivers, streams, etc.)

eg ````curl "http://mapp.life:8000/nearme?x=-111.45&y=45.567&f=json&type=poi" -o nearme.txt````
## DATA conversion endpoint:  serveraddress/data/  [multipart file]

Hit this enpoint with a multipart file, one is the map of the data (info) the other is the file itself (file) and it'll give back a converted dataset prepared for Unity / DeepAR according to json structure in config/

#### Part 1  "info" []byte of info array
https://github.com/lumin8/deepar-data/blob/master/config/input.json

#### Part 2  "file" []byte of file


## Current Structure of OUTBOUND Data
-Singlepart File

#### Json String
https://github.com/lumin8/deepar-data/blob/master/config/output.json


## Tests
One test currently exists as 'data_test.go'.  This can be run from any machine, it bundles up a CSV from tests/trek/ and the json array of info (above), hits the endpoint, and receives the data back.  Currently, a 200KB csv and a nominal json expand **3x** in the current arrangement for datasets (config/output.json)... this is not ideal, further testing may indicate a need to compress the verbosity of the output.json.

## TBD
Add SHP conversion functionality
Add DXF conversion functionality
Handle different Coordinate Systems

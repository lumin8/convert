# CONVERT... csv/shp/dxf to custom unity JSON API

This api takes http POST of csv (tbd shp and dxf), and returns json or a struct used to build Unity objects for the Deep AR project.

main.go farms out requests to convert / etc.
datamap.go holds the structs
convert.go is the workhorse: it peels into the data,  _adds elevations to coordinates_, and spits out a custom json for use with unity mobile app.
elevations.go is the sibling to convert.go, there as a tool that performs the elevation lookups

**convert.go relies on a huge digital elevation model of the earth (several hundred gb's of data)**

- convert.go does _not_ currently download this dataset
- convert.go _should_ be set up to download the dataset (eg. in docker, etc)
- convert.go _does_ rely on the **gdal** tools to request point specific data.

Logging is only currently available if one runs the datapi with NOHUP

## TBD 

- change struct of inbound/outbound data to absorb json, and get/put from gs bucket, INSTEAD OF multipart file
- should prolly wrap this in a container, se we can have:
- local installation of gdal tools
- upgrade makefile to pull and untar ~94 gb of DEM data from `gs://data.map.life/raw/dem/worlddem_100m.tar.gz`, if not already exists
- remove main.go if this is not being used as standalone service (all main.go does is set up server mux)


## SAMPLE endpoint:  .../sample?get=&did=&token=" -o testsample.json

Get Options:
- **dataset**
- **collection**

Datasets Ids ( DID ) currently supported:
- **1**
- **2**
- **3**

eg: ````curl "http://data.map.life/sample?get=collection" -o samplecollection.json````

## DATA conversion endpoint:  .../convert/  [multipart file]

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

## TBD Future Conversion Functionality
Add SHP conversion
Add DXF conversion
Handle different Coordinate Systems
Handle input of EPSG:3857 (utm meters) in addition to the already-accepted EPSG:4326 (lat long)

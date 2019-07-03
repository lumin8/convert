# CONVERT... csv/geojson to custom unity json format for APP

This api takes http POST of csv and geojson, and returns struct used to build Unity objects for the Deep AR project.

convert.go functionality includes: parsing inbound data,  _adds elevations to coordinates_, ensures output is in EPSG:3857 projection, finds center point x,y,z, and spits out a custom json for use with unity mobile app.

**convert.go relies on a huge digital elevation model of the earth (several hundred gb's of data)**

- convert.go does _not_ currently download this dataset
- convert.go _should_ be set up to download the dataset (eg. in docker, etc)
- convert.go _does_ rely on the **gdal** tools to request point specific data.

## TBD 

- should prolly wrap this in a container, se we can have:
- local installation of gdal tools
- upgrade makefile to pull and untar ~94 gb of DEM data from `gs://data.map.life/raw/dem/worlddem_100m.tar.gz`, if not already exists


## DATA conversion endpoint:  .../convert/  [multipart file]

Hit this enpoint with a multipart file, one is the map of the data (info) the other is the file itself (file) and it'll give back a converted dataset prepared for Unity / DeepAR according to json structure in config/

## Current Structure of OUTBOUND Data
-Singlepart File

## Tests
One test currently exists as 'data_test.go'.  This can be run from any machine, it bundles up a CSV from tests/trek/ and the json array of info (above), hits the endpoint, and receives the data back.  Currently, a 200KB csv and a nominal json expand **3x** in the current arrangement for datasets (config/output.json)... this is not ideal, further testing may indicate a need to compress the verbosity of the output.json.

## TBD Future Conversion Functionality
Add SHP conversion
Add DXF conversion
Handle different Coordinate Systems

# Data
It's all about the data.

Currently, the DeepAR project uses:
- json  for points, surfaces (Digital Elevation Models 'DEM'), lines?
- raw   for surfaces (in progress)

### X Y Z ... world vs unity axis
When describing position on/in the earth,
X refers to *easting* (how far east/west you are from the Prime Meridian)
Y refers to *northing* (how far north/south you are from the Equator)
Z refers to *elevation* (how far above/below the ground)

When describing position on/in unity,
X refers to *easting*
Z refers to *northing*
Y refers to *elevation*

Take careful note of the Y/Z switch!!

### Spatial Reference System 'SRS' (also called 'ESPG')

Clients should ubiquitously use the data in EPSG:3857 format
- standardize all data coming in / out of the platform
- Google, Bing, the rest of the world ises this projection
- easy to convert server-side

VERY important information about EPSG: 3857
- any points on the earth WEST of the Prime Meridian (England) are _negative_
- any points south of the equater are _negative_


## Tools
Several scripts are written in python (quick and dirty) to facilitate the conversion of data for the unity platform.  These scripts include:

csv2json.py
tif2json.py
tif2raw.py

Visit the /tools directory for detailed information on their use.

## Data
The provided dataset is of the 'Trek' project in Northern B.C, Canada.  Steep terrain, in meters
Includes:
- points (drill hole intercepts)
- surface 
    --as json (a cloud of points)
    --as DEM (in .raw format)

### YML: how DeepAR stores metadata regarding all datasets
All datasets, when the API is operational, should prescribe a YML file (via use input) that maps the following, of any incoming data:
-x_field
-y_field
-z_field
-srs
-meters
See the yml files in the Trek sample folder for examples

### TBD
- load up json data as lines
- continue testing surfaces as json point clouds (see *heightmap* in the Trek sample data)
- build up attributes for use in the colorizing of the featuresi
- extend the YML files to include username/id, date, etc metadata
- add an azimuth to points, if they have it, so they can be made into cylinders

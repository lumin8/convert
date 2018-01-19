# Data
It's all about the data.

Currently, the DeepAR project uses:
- json  for points, surfaces (Digital Elevation Models 'DEM'), lines?
- raw   for surfaces (in progress)


## Sample
Dataset is of the 'Trek' project in Northern B.C, Canada.  Steep terrain, in meters

## Spatial Reference System 'SRS' or 'ESPG'

Clients should ubiquitously use the data in EPSG:3857 format
- standardize all data coming in / out of the platform
- Google, Bing, the rest of the world ises this projection
- easy to convert server-side

## YML
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
- build up attributes for use in the colorizing of the features
- add an azimuth to points, if the have it, so they can be made into cylinders


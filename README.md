# Convert

Convert holds several golang tools used to convert csv, postgres json, and geojson into MineAR (Unity) - style json.

This tool is a replicate of the pkg by same name within MineAR's **admin** utility, and heavily used by Map.Life's **dataman** utility.

The output is a ```Datasets``` struct, which can hold any number of features (points, lines, and shapes), and attributes for each feature.

The final ```Datasets``` struct must be json-marshaled prior to use in MineAR.

Note: this package spawns a unique channel & goroutine for each dataset processed, called an `ExtentContainer`.  The purpose of this `ExtentContainer` is to asynchronously handle coordinates flying into the channel, figure out which four form the bottom-left and top-right coordinates of the enclosing bounding box `bbox` (aka *Extent*), which is then processed to find the Center point, to which a user will zoom if they select the `Datasets` in VR mode in MineAR.  This `bbox` also forms the basis for configuring the S2 coverage- a series of keys or 'tokens' representing certain polygons on the ground- which may be used to find associated features or datasets (eg. roads, rivers, DEM, etc) from key:value stores without having to perform complicated `ST_Intersects` queries.   The bbox and S2 keys are very important.


## Primary Functions

### DatasetFromCSV(xField string, yField string, zField string, contents io.Reader) (*Datasets, error)
Converts a csv file (with x, y, and z fields (if known) into a `Datasets` struct.


### DatasetFromGEOJSON(xField string, yField string, zField string, contents io.Reader) (*Datasets, error)
Converts a regular GEOJSON (x, y, and z fields are useless, geojson has rules on these field names) into a `Datasets` struct.


### parseGEOJSONCollection(collection *geojson.FeatureCollection, container *ExtentContainer) (*Datasets, error)
Takes a GEOJSON- which are always 'feature collections', and breaks it up into features.  Depends on ParseGEOJSONFeature.  Uses the intermediate `FeatureInfo` struct as a map between the geojson itself and the new `Datasets` which holds n count of Features of type `FeatureInfo`.
You should not call this function directly, but rather DatasetFromGEOJSON or, if you have individual features, ParseGEOJSONFeature.


### ParseGEOJSONFeature(gfeature *convert.FeatureInfo, outdataset *convert.Datasets, nil)```
Converts each feature json into a `FeatureInfo` class, parsing both the attributes of the originating geojson and the geom of the originating geom.


### ParseGEOJSONGeom
Explodes the feature geometry, coordinate by coordinate, uses `GetElev` to fill in Z value if needed, and enforces EPSG:3857.


### ParseGEOJSONAttributes
Explodes the feature attributes, maps *name*, *styletype*, and *id* to a higher object level in the `FeatureInfo`, removes attributes with missing/nil values (keeping the resulting Unity json as trim as possible), and moves all cleaned key:value attribute pairs to the new `FeatureInfo`.


## Secondary Functions  (these should be moved to seperate toolset TBD!!!)

### GetElev(x float64, y float64) (float64, error)
Takes x and y, provides a single elevation in *meters*
Depends upon `To4326`

### To4326(x float64, y float64) (float64, float64)
Takes x and y, provides x and y in EPSG:4326  (lat lon decimal)

### To3857((x float64, y float64) (float64, float64)
Takes x and y, provides x and y in EPSG:3847  (universal web mercator)



# csv2json.py
Converts a csv to unity json

`python xyz2json.py /some/file.csv`

- requires a corresponding yml file that maps headers to x, y, z world axis, even if columns in the csv are actually named x, y and z
- defines a geometry type of 
-- points
-- lines
-- polygons

# tif2json.py
Converts a geotif to a unity json 'cloud of surface points'

`python tif2json.py /some/file.tif`

- requires a corresponding yml file that maps spatial reference system (srs) of original tif
- converts to 3857 (temp file, removed after processing .raw file)

# tif2raw.py
Converts a geotif to a .RAW file for unity import

`python tif2raw.py /some/file.tif`

- requires a corresponding yml file that maps spatial reference system (srs) of original tif
- converts to 3857 (temp file, removed after processing .raw file)

# unity json format 
```
{
  "points": [
    {
      "x": -14615548,
      "y": 7772618,
      "z": 1298,
      "foo": ...
      "bar" : ...
    }
  ]
}
```

# known TBD
- json formats exported by this tool are not formatted with \n or spaces!!!

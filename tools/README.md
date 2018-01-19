# csv2json.py

## usage

`python xyz2json.py /some/file.csv`

- requires a corresponding yml file that maps headers to x, y, z world axis, even if columns in the csv are actually named x, y and z
- defines a geometry type of 
-- points
-- lines
-- polygons

converts a csv to a unity json

# xyz2json.py

## usage

`python xyz2json.py /some/file.tif`

converts a tif to a unity json

# unity json format 
```{
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

# known TBD
- json formats exported by this tool are not formatted with \n or spaces!!!

package main

const (
    Convert = "tools/convert.py" //tbd handle csv, shp dem...
    Getdem = "tools/dem.py"
)

type Projects struct {
    Id string `json:"id" yaml:"id"`
    Name string `json:"name" yaml:"name"`
    Updated string `json:"lastUpdated" yaml:"lastUpdated"`
    S2hash string `json:"s2hash" yaml:"s2hash"`
    Dem []Dem
    Datasets []Datasets
}

type Dem struct {
    Id string `json:"id" yaml:"id"`
    Name string `json:"name" yaml:"name"`
    Updated string `json:"lastUpdated" yaml:"lastUpdated"`
    S2hash string `json:"s2hash" yaml:"s2hash"`
    Points []Points
}

type Datasets struct {
    Id string `json:"id" yaml:"id"`
    Name string `json:"name" yaml:"name"`
    Url string `json:"dataurl" yaml:"dataurl"`
    Updated string `json:"lastUpdated" yaml:"lastUpdated"`
    Center []Center
    Bbox string `json:"bbox" yaml:"bbox"`
    S2hash string `json:"id" yaml:"id"`
    Xfield string `json:"xfield" yaml:"xfield"`
    Yfield string `json:"yfield" yaml:"yfield"`
    Zfield string `json:"zfield" yaml:"zfield"`
    Points []Points
    Lines []Lines
    Shapes []Shapes
    Format string `json:"format" yaml:"format"`
}

type Center struct {
    X int `json:"x" yaml:"x"`
    Y int `json:"y" yaml:"y"`
    Z int `json:"z" yaml:"z"`
}

type Points struct {
    X int `json:"x" yaml:"x"`
    Y int `json:"y" yaml:"y"`
    Z int `json:"z" yaml:"z"`
    Attributes []Attributes
}

type Lines struct {
    X int `json:"x" yaml:"x"`
    Y int `json:"y" yaml:"y"`
    Z int `json:"z" yaml:"z"`
    Attributes []Attributes
}

type Shapes struct {
    X int `json:"x" yaml:"x"`
    Y int `json:"y" yaml:"y"`
    Z int `json:"z" yaml:"z"`
    Attributes []Attributes
}

type Attributes struct {
    Key string `json:"key" yaml:"key"`
    Value string `json:"value" yaml:"value"`
}

type Input struct {
    Id int `json:"id" yaml:"id"`
    Xfield string `json:"xfield" yaml:"xfield"`
    Yfield string `json:"yfield" yaml:"yfield"`
    Zfield string `json:"zfield" yaml:"zfield"`
    Srs int `json:"srs" yaml:"srs"`
    Units string `json:"units" yaml:"units"`
    Format string `json:"format" yaml:"format"`
}

type Tools struct {
    Pytools []Pytools
    Demfiles []Demfiles
}

type Pytools struct {
    CsvConv string `json:"csv" yaml:"csv"`
    ShpConv string `json:"shp" yaml:"shp"`
    DxfConv string `json:"dxf" yaml:"dxf"`
}

type Demfiles struct {
    Usa string `json:"usa" yaml:"usa"`
    World string `json:"world" yaml:"world"`
}

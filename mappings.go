package main

const (
    Convert = "tools/convert.py"
    Getdem = "tools/dem.py"
)


type Project struct {
    Id string `json:"id" yaml:"id"`
    Dem []Dem
    Datasets []Datasets
}

type Dem struct {
    Id string `json:"id" yaml:"id"`
    Points []Points
}

type Datasets struct {
    Id string `json:"id" yaml:"id"`
    Url string `json:"dataurl" yaml:"dataurl"`
    Updated string `json:"lastUpdated" yaml:"lastUpdated"`
    Center []Center
    Bbox string `json:"bbox" yaml:"bbox"`
    S2hash string `json:"id" yaml:"id"`
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
    Id string `json:"dataset" yaml:"dataset"`
    Xfield float64 `json:"xfield" yaml:"xfield"`
    Yfield float64 `json:"yfield" yaml:"yfield"`
    Zfield float64 `json:"zfield" yaml:"zfield"`
    Srs int `json:"srs" yaml:"srs"`
    Geom string `json:"geom" yaml:"geom"`
    Units string `json:"units" yaml:"units"`
    Format string `json:"format" yaml:"format"`
    Data string `json:"data" yaml:"data"`
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

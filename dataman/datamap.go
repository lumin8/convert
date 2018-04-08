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
    Points []([]float64)
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
    X float64 `json:"x" yaml:"x"`
    Y float64 `json:"y" yaml:"y"`
    Z float64 `json:"z" yaml:"z"`
}

type DemPoints struct {
    Point
}

type Point struct {
    Point []float64
}

//type DemPoints struct {
//    X float64 `json:"x" yaml:"x"`
//    Y float64 `json:"y" yaml:"y"`
//    Z float64 `json:"z" yaml:"z"`
//}

type Points struct {
    X float64 `json:"x" yaml:"x"`
    Y float64 `json:"y" yaml:"y"`
    Z float64 `json:"z" yaml:"z"`
    Attributes []Attributes
}

type Lines struct {
    X float64 `json:"x" yaml:"x"`
    Y float64 `json:"y" yaml:"y"`
    Z float64 `json:"z" yaml:"z"`
    Attributes []Attributes
}

type Shapes struct {
    X float64 `json:"x" yaml:"x"`
    Y float64 `json:"y" yaml:"y"`
    Z float64 `json:"z" yaml:"z"`
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



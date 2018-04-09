package main

type Projects struct {
    Id string `json:"id" yaml:"id"`
    Name string `json:"name" yaml:"name"`
    Updated string `json:"lastUpdated" yaml:"lastUpdated"`
    S2hash string `json:"s2hash" yaml:"s2hash"`
    Dem []Dem
    Datasets []Datasets
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

type Dem struct {
    Id string `json:"id" yaml:"id"`
    Name string `json:"name" yaml:"name"`
    Updated string `json:"lastUpdated" yaml:"lastUpdated"`
    S2hash string `json:"s2hash" yaml:"s2hash"`
    Points []([]float64)
}

type Points struct {
    Point ([]float64)
    Attributes []Attributes
}

type Lines struct {
    Id string `json:"id" yaml:"id"`
    Name string `json:"name" yaml:"name"`
    Attributes []Attributes
    Points []([]float64)
}

type Shapes struct {
    Id string `json:"id" yaml:"id"`
    Name string `json:"name" yaml:"name"`
    Attributes []Attributes
    Points []([]float64)
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

type Geojson struct {
    Type string `json:"type" yaml:"type"`
    Coords []([]float64)
}


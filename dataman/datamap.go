package main

type Projects struct {
    Id string `json:"id" yaml:"id"`
    Name string `json:"name" yaml:"name"`
    Updated string `json:"lastUpdated" yaml:"lastUpdated"`
    S2hash string `json:"s2hash" yaml:"s2hash"`
    Dem []Dem `json:"dem" yaml:"dem"`
    Datasets []Datasets `json:"datasets" yaml:"datasets"`
}

type Datasets struct {
    Id string `json:"id" yaml:"id"`
    Name string `json:"name" yaml:"name"`
    Url string `json:"dataurl" yaml:"dataurl"`
    Updated string `json:"lastUpdated" yaml:"lastUpdated"`
    Center []Point `json:"center" yaml:"center"`
    Bbox string `json:"bbox" yaml:"bbox"`
    S2hash string `json:"id" yaml:"id"`
    Points []Points `json:"points" yaml:"points"`
    Lines []Lines `json:"lines" yaml:"lines"`
    Shapes []Shapes `json:"shapes" yaml:"shapes"`
}

type Point struct {
    X float64 `json:"x" yaml:"x"`
    Y float64 `json:"y" yaml:"y"`
    Z float64 `json:"z" yaml:"z"`
}

type Pointarray struct {
    Points []([]float64) `json:"points" yaml:"points"`
}

type Dem struct {
    Id string `json:"id" yaml:"id"`
    Name string `json:"name" yaml:"name"`
    Updated string `json:"lastUpdated" yaml:"lastUpdated"`
    S2hash string `json:"s2hash" yaml:"s2hash"`
    Points []([]float64) `json:"points" yaml:"points"`
}

type Points struct {
    Id string `json:"id" yaml:"id"`
    Name string `json:"name" yaml:"name"`
    Attributes []Attributes `json:"attributes" yaml:"attributes"`
    Point []float64 `json:"point" yaml:"point"`
}

type Lines struct {
    Id string `json:"id" yaml:"id"`
    Name string `json:"name" yaml:"name"`
    Attributes []Attributes `json:"attributes" yaml:"attributes"`
    Points []([]float64) `json:"points" yaml:"points"`
}

type Shapes struct {
    Id string `json:"id" yaml:"id"`
    Name string `json:"name" yaml:"name"`
    Attributes []Attributes `json:"attributes" yaml:"attributes"`
    Points []([]float64) `json:"points" yaml:"points"`
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

type GeojsonS struct {
    Type string `json:"type" yaml:"type"`
    Coords []float64 `json:"coordinates" yaml:"coordinates"`
}

type GeojsonM struct {
    Type string `json:"type" yaml:"type"`
    Coords []([]float64) `json:"coordinates" yaml:"coordinates"`
}


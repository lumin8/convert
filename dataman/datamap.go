package main

import (
    "github.com/paulmach/go.geojson"
    "github.com/golang/geo/s2"
)

type Projects struct {
    Id string `json:"id" yaml:"id"`
    Name string `json:"name" yaml:"name"`
    Updated string `json:"lastUpdated" yaml:"lastUpdated"`
    S2 []string `json:"s2" yaml:"s2"`
    Dem []Dem `json:"dem" yaml:"dem"`
    Datasets []Datasets `json:"datasets" yaml:"datasets"`
}

type Datasets struct {
    Id string `json:"id" yaml:"id"`
    Name string `json:"name" yaml:"name"`
    Url string `json:"dataurl" yaml:"dataurl"`
    Updated string `json:"lastUpdated" yaml:"lastUpdated"`
    Center []Point `json:"center" yaml:"center"`
    S2 []string `json:"s2" yaml:"s2"`
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
    S2 string `json:"s2" yaml:"s2"`
    Points []([]float64) `json:"points" yaml:"points"`
}

type Points struct {
    Id string `json:"id" yaml:"id"`
    Name string `json:"name" yaml:"name"`
    StyleType string `json:"type" yaml:"type"`
//    Attributes []map[string]interface{} `json:"attributes" yaml:"attributes"`
    Attributes []Attributes `json:"attributes" yaml:"attributes"`
    Point []float64 `json:"point" yaml:"point"`
}

type Lines struct {
    Id string `json:"id" yaml:"id"`
    Name string `json:"name" yaml:"name"`
    StyleType string `json:"type" yaml:"type"`
//    Attributes []map[string]interface{} `json:"attributes" yaml:"attributes"`
    Attributes []Attributes `json:"attributes" yaml:"attributes"`
    Points []([]float64) `json:"points" yaml:"points"`
}

type Shapes struct {
    Id string `json:"id" yaml:"id"`
    Name string `json:"name" yaml:"name"`
    StyleType string `json:"type" yaml:"type"`
//    Attributes []map[string]interface{} `json:"attributes" yaml:"attributes"`
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
    Inurl string `json:"inurl" yaml:"inurl"`
    Outurl string `json:"outurl" yaml:"outurl"`
}

type GeojsonS struct {
    Type string `json:"type" yaml:"type"`
    Coords []float64 `json:"coordinates" yaml:"coordinates"`
}

type GeojsonM struct {
    Type string `json:"type" yaml:"type"`
    Coords []([]float64) `json:"coordinates" yaml:"coordinates"`
}

type FeatureInfo struct {
    id		int
    geojson     geojson.Feature //G
    geomtype    string
    srid        string
    s2          []s2.CellID
    tokens      []string
    name	string
    styletype	string
}

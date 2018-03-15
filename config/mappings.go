type Project struct {
    Id string `yaml:"id"`
    Dem []Dem
    Datasets []Datasets
}

type Dem struct {
    Id string `yaml:"id"`
    Points []Points
}

type Datasets struct {
    Id string `yaml:"id"`
    Center []Center
    Bbox string `yaml:"bbox"`
    S2hash string `yaml:"id"`
    Points []Points
    Lines []Lines
    Shapes []Shapes
}

type Center struct {
    X int `yaml:"x"`
    Y int `yaml:"y"`
    Z int `yaml:"z"`
}

type Points struct {
    X int `yaml:"x"`
    Y int `yaml:"y"`
    Z int `yaml:"z"`
    Attributes []Attributes
}

type Attributes struct {
    Key string `yaml:"key"`
    Value string `yaml:"value"`
}

type Input struct {
    Id string `yaml:"dataset"`
    Xfield float64 `yaml:"xfield"`
    Yfield float64 `yaml:"yfield"`
    Zfield float64 `yaml:"zfield"`
    Srs int `yaml:"srs"`
    Geom string `yaml:"geom"`
    Units string `yaml:"units"`
    Format string `yaml:"format"`
    Data string `yaml:"data"`
}

type Tools struct {
    Pytools []Pytools
    Demfiles []Demfiles
}

type Pytools struct {
    CsvConv string `yaml:"csv"`
    ShpConv string `yaml:"shp"`
    DxfConv string `yaml:"dxf"`
}

type Demfiles struct {
    Usa string `yaml:"usa"`
    World string `yaml:"world"`
}

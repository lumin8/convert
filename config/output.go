package config

type Projects struct {
	Dataset []struct {
		ID          interface{} `json:"id"`
		Name        interface{} `json:"name"`
		LastUpdated interface{} `json:"lastUpdated"`
		URL         interface{} `json:"url"`
		Center      struct {
			X interface{} `json:"x"`
			Y interface{} `json:"y"`
			Z interface{} `json:"z"`
		} `json:"center"`
		Bbox   interface{} `json:"bbox"`
		S2Hash interface{} `json:"s2hash"`
		Points []struct {
			X          interface{} `json:"x"`
			Y          interface{} `json:"y"`
			Z          interface{} `json:"z"`
			Attributes []struct {
				Key   interface{} `json:"key"`
				Value interface{} `json:"value"`
			} `json:"attributes"`
		} `json:"points"`
		Lines []struct {
			ID         interface{} `json:"id"`
			Attributes []struct {
				Key   interface{} `json:"key"`
				Value interface{} `json:"value"`
			} `json:"attributes"`
			Points []struct {
				X interface{} `json:"x"`
				Y interface{} `json:"y"`
				Z interface{} `json:"z"`
			} `json:"points"`
		} `json:"lines"`
		Shapes []struct {
			ID         interface{} `json:"id"`
			Attributes []struct {
				Key   interface{} `json:"key"`
				Value interface{} `json:"value"`
			} `json:"attributes"`
			Points []struct {
				X interface{} `json:"x"`
				Y interface{} `json:"y"`
				Z interface{} `json:"z"`
			} `json:"points"`
		} `json:"shapes"`
	} `json:"dataset"`
}

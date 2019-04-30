package volkszaehler

import "encoding/json"

// EntityResponse is the middleware response to /entity.json
type EntityResponse struct {
	Version  string   `json:"version"`
	Entities []Entity `json:"entities"`
}

// Entity is a single middleware entity
type Entity struct {
	UUID     string   `json:"uuid"`
	Type     string   `json:"type"`
	Title    string   `json:"title"`
	Children []Entity `json:"children"`
}

// DataResponse is the middleware response to /data.json
type DataResponse struct {
	Version string      `json:"version"`
	Data    Data        `json:"data"`
	Debug   interface{} `json:"debug"`
}

// Data holds the array of middleware tuples
type Data struct {
	Tuples []Tuple `json:"tuples"`
}

// Tuple is a single timestamp/value tuple
type Tuple struct {
	Timestamp int64
	Value     float32
}

type PrognosisResponse struct {
	Version   string    `json:"version"`
	Prognosis Prognosis `json:"prognosis"`
}

type Prognosis struct {
	Consumption float32 `json:"consumption"`
	Fator       float32 `json:"factor"`
}

// UnmarshalJSON converts volkszaehler tuple into Tuple struct
func (t *Tuple) UnmarshalJSON(b []byte) error {
	var a []*json.RawMessage
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}

	if err := json.Unmarshal(*a[0], &t.Timestamp); err != nil {
		return err
	}

	if err := json.Unmarshal(*a[1], &t.Value); err != nil {
		return err
	}

	return nil
}

package volkszaehler

import "encoding/json"

// EntityType represent the entity types enum
type EntityType string

const (
	// Channel is the Channel entity type
	Channel EntityType = "channel"
	// Group is the Aggregator entity type
	Group EntityType = "group"
)

// EntitiesResponse is the middleware response to /entity.json
type EntitiesResponse struct {
	Version   string    `json:"version"`
	Exception Exception `json:"exception"`
	Entities  []Entity  `json:"entities"`
}

// EntityResponse is the middleware response to /entity/uuid.json
type EntityResponse struct {
	Version   string    `json:"version"`
	Exception Exception `json:"exception"`
	Entity    Entity    `json:"entity"`
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
	Version   string      `json:"version"`
	Exception Exception   `json:"exception"`
	Data      Data        `json:"data"`
	Debug     interface{} `json:"debug"`
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

// PrognosisResponse is the middleware response to /prognosis.json
type PrognosisResponse struct {
	Version   string    `json:"version"`
	Exception Exception `json:"exception"`
	Prognosis Prognosis `json:"prognosis"`
}

// Prognosis is the prognosis result
type Prognosis struct {
	Consumption float32 `json:"consumption"`
	Factor      float32 `json:"factor"`
}

// Exception is the middleware exception structure
type Exception struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// PostDataResponse is the middleware response to POST requests to /data.json
type PostDataResponse struct {
	Version   string    `json:"version"`
	Exception Exception `json:"exception"`
	Rows      int       `json:"rows"`
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

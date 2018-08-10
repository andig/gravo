package main

import "encoding/json"

type EntityResponse struct {
	Version  string   `json:"version"`
	Entities []Entity `json:"entities"`
}

type Entity struct {
	UUID     string   `json:"uuid"`
	Type     string   `json:"type"`
	Title    string   `json:"title"`
	Children []Entity `json:"children"`
}

type DataResponse struct {
	Version string      `json:"version"`
	Data    DataStruct  `json:"data"`
	Debug   interface{} `json:"debug"`
}

type DataStruct struct {
	Tuples []Tuple `json:"tuples"`
}

type Tuple struct {
	Timestamp int64
	Value     float32
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

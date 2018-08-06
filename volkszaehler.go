package main

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

type Tuple []float64

package main

import "time"

// AnnotationsRequest encodes the information provided by Grafana in /annotations.
type AnnotationsRequest struct {
	Range      Range         `json:"range"`
	RangeRaw   RelativeRange `json:"rangeRaw"`
	Annotation Annotation    `json:"annotation"`
}

// AnnotationResponse describes an annotation event
// https://github.com/grafana/simple-json-datasource#annotation-api
type AnnotationResponse struct {
	// The original annotation sent from Grafana.
	Annotation Annotation `json:"annotation"`
	// Time since UNIX Epoch in milliseconds. (required)
	Time int64 `json:"time"`
	// The title for the annotation tooltip. (required)
	Title string `json:"title"`
	// Tags for the annotation. (optional)
	Tags string `json:"tags"`
	// Text for the annotation. (optional)
	Text string `json:"text"`
}

type RelativeRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// Range specifies the time range the request is valid for.
type Range struct {
	From time.Time     `json:"from"`
	To   time.Time     `json:"to"`
	Raw  RelativeRange `json:"raw"`
}

// Annotation is the object passed by Grafana when it fetches annotations.
// http://docs.grafana.org/plugins/developing/datasources/#annotation-query
type Annotation struct {
	// Name must match in the request and response
	Name string `json:"name"`

	Datasource string `json:"datasource"`
	IconColor  string `json:"iconColor"`
	Enable     bool   `json:"enable"`
	ShowLine   bool   `json:"showLine"`
	Query      string `json:"query"`
}

// TagKeyResponse encodes additional query options
type TagKeyResponse struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// TagValueResponse encodes additional query option values
type TagValueResponse struct {
	Text string `json:"text"`
}

// QueryRequest encodes the information provided by Grafana in /query.
// https://github.com/grafana/simple-json-datasource#query-api
type QueryRequest struct {
	PanelID       int           `json:"panelId"`
	Range         Range         `json:"range"`
	RangeRaw      RelativeRange `json:"rangeRaw"`
	Interval      string        `json:"interval"`
	IntervalMs    int           `json:"intervalMs"`
	Targets       []Target      `json:"targets"`
	AdhocFilters  []Filter      `json:"adhocFilters"`
	Format        string        `json:"json"`
	MaxDataPoints int           `json:"maxDataPoints"`
}

// QueryResponse contains information to render query result.
type QueryResponse struct {
	Target     interface{} `json:"target"`
	Datapoints []Tuple     `json:"datapoints"`
}

// Target describes a query target
type Target struct {
	Target string     `json:"target"`
	RefID  string     `json:"refId"`
	Type   string     `json:"type"`
	Data   TargetData `json:"data,omitempty"`
}

type TargetData map[string]string

// Filter is a compontent of adhoc filters
type Filter struct {
	Key      string `json:"key"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

// SearchRequest encodes the information provided by Grafana in /query.
// https://github.com/grafana/simple-json-datasource#search-api
type SearchRequest struct {
	Target string `json:"target"`
}

// SearchResponse contains information to render search result.
type SearchResponse struct {
	Text string `json:"text"`
	UUID string `json:"value"`
}

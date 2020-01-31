package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/andig/gravo/grafana"
	"github.com/andig/gravo/volkszaehler"
)

// Server is the http endpoint used by Grafana's SimpleJson plugin
type Server struct {
	api         volkszaehler.Client
	cacheMux    sync.Mutex // guards entityCache
	entityCache map[string]string
}

func newServer(api volkszaehler.Client) *Server {
	server := &Server{
		api:         api,
		entityCache: make(map[string]string),
	}

	// get entity map on startup
	server.getPublicEntites()

	return server
}

func (server *Server) rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok\n")
}

func (server *Server) annotationsHandler(w http.ResponseWriter, r *http.Request) {
	ar := grafana.AnnotationsRequest{}
	if err := json.NewDecoder(r.Body).Decode(&ar); err != nil {
		http.Error(w, fmt.Sprintf("json decode failed: %v", err), http.StatusBadRequest)

		return
	}

	resp := []grafana.AnnotationResponse{}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("json encode failed: %v", err)
		http.Error(w, fmt.Sprintf("json encode failed: %v", err), http.StatusInternalServerError)

		return
	}
}

func (server *Server) tagKeysHandler(w http.ResponseWriter, r *http.Request) {
	resp := []grafana.TagKeyResponse{
		grafana.TagKeyResponse{
			Type: "string",
			Text: "group"},
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("json encode failed: %v", err)
		http.Error(w, fmt.Sprintf("json encode failed: %v", err), http.StatusInternalServerError)

		return
	}
}

func (server *Server) tagValuesHandler(w http.ResponseWriter, r *http.Request) {
	resp := []grafana.TagValueResponse{
		grafana.TagValueResponse{Text: "Current"},
		grafana.TagValueResponse{Text: "Consumption"},
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("json encode failed: %v", err)
		http.Error(w, fmt.Sprintf("json encode failed: %v", err), http.StatusInternalServerError)

		return
	}
}

func (server *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	sr := grafana.SearchRequest{}
	if err := json.NewDecoder(r.Body).Decode(&sr); err != nil {
		log.Printf("json decode failed: %v", err)
		http.Error(w, fmt.Sprintf("json decode failed: %v", err), http.StatusBadRequest)

		return
	}

	resp := server.executeSearch()

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("json encode failed: %v", err)
		http.Error(w, fmt.Sprintf("json encode failed: %v", err), http.StatusInternalServerError)

		return
	}
}

func (server *Server) flattenEntities(result *[]volkszaehler.Entity, entities []volkszaehler.Entity, parent string) {
	for _, entity := range entities {
		if entity.Type == "group" {
			server.flattenEntities(result, entity.Children, entity.Title)
		} else {
			if parent != "" {
				entity.Title = fmt.Sprintf("%s (%s)", entity.Title, parent)
			}
			*result = append(*result, entity)
		}
	}
}

func (server *Server) populateCache(entities []volkszaehler.Entity) {
	server.cacheMux.Lock()
	defer server.cacheMux.Unlock()

	if len(entities) > 0 {
		server.entityCache = make(map[string]string)
	}

	// add to cache
	for _, entity := range entities {
		if _, ok := server.entityCache[entity.UUID]; !ok {
			server.entityCache[entity.UUID] = entity.Title
		}
	}
}

func (server *Server) getPublicEntites() []volkszaehler.Entity {
	entities := make([]volkszaehler.Entity, 0)

	publicEntities, err := server.api.QueryPublicEntities()
	if err != nil {
		log.Printf("api call failed: %v", err)

		return entities
	}

	server.flattenEntities(&entities, publicEntities, "")
	server.populateCache(entities)

	return entities
}

func (server *Server) executeSearch() []grafana.SearchResponse {
	entities := server.getPublicEntites()

	res := []grafana.SearchResponse{}
	for _, entity := range entities {
		res = append(res, grafana.SearchResponse{
			Text: entity.Title,
			UUID: entity.UUID,
		})
	}

	return res
}

func (server *Server) queryHandler(w http.ResponseWriter, r *http.Request) {
	qr := grafana.QueryRequest{}
	if err := json.NewDecoder(r.Body).Decode(&qr); err != nil {
		log.Printf("json decode failed: %v", err)
		http.Error(w, fmt.Sprintf("json decode failed: %v", err), http.StatusBadRequest)

		return
	}

	resp := server.executeQuery(qr)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("json encode failed: %v", err)
		http.Error(w, fmt.Sprintf("json encode failed: %v", err), http.StatusInternalServerError)

		return
	}
}

func roundTimestampMS(ts int64, group string) int64 {
	const millisPerSecond = 1000
	t := time.Unix(ts/millisPerSecond, 0)

	switch group {
	case "hour":
		t.Truncate(time.Hour)
	case "day":
		t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
	case "month":
		t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.Local)
	}

	return t.Unix() * 1000
}

func (server *Server) executeQuery(qr grafana.QueryRequest) []grafana.QueryResponse {
	res := make([]grafana.QueryResponse, len(qr.Targets))
	wg := &sync.WaitGroup{}

	for idx, target := range qr.Targets {
		wg.Add(1)

		go func(idx int, target grafana.Target) {
			var qres grafana.QueryResponse

			context := strings.ToLower(target.Data.Context)
			if context == "prognosis" {
				qres = server.queryPrognosis(target)
			} else {
				qres = server.queryData(target, &qr)
			}

			// substitute name
			server.cacheMux.Lock()
			if text, ok := server.entityCache[qres.Target.(string)]; ok {
				qres.Target = text
			}
			server.cacheMux.Unlock()

			if target.Data.Name != "" {
				qres.Target = target.Data.Name
			}

			res[idx] = qres

			wg.Done()
		}(idx, target)
	}

	wg.Wait()

	return res
}

func (server *Server) queryData(target grafana.Target, qr *grafana.QueryRequest) grafana.QueryResponse {
	qres := grafana.QueryResponse{
		Target:     target.Target,
		Datapoints: []grafana.ResponseTuple{},
	}

	var group string
	if target.Data.Group != "" {
		group = strings.ToLower(target.Data.Group)
	}

	var options string
	if target.Data.Options != "" {
		options = strings.ToLower(target.Data.Options)
	}

	tuples, err := server.api.QueryData(
		target.Target,
		qr.Range.From,
		qr.Range.To,
		group,
		options,
		qr.MaxDataPoints,
	)
	if err != nil {
		log.Printf("api call failed: %v", err)
		return qres
	}

	for _, tuple := range tuples {
		if group != "" {
			tuple.Timestamp = roundTimestampMS(tuple.Timestamp, group)
		}

		qres.Datapoints = append(qres.Datapoints, grafana.ResponseTuple{
			Timestamp: tuple.Timestamp,
			Value:     tuple.Value,
		})
	}

	return qres
}

func (server *Server) queryPrognosis(target grafana.Target) grafana.QueryResponse {
	qres := grafana.QueryResponse{
		Target:     target.Target,
		Datapoints: []grafana.ResponseTuple{},
	}

	if target.Data.Period != "" {
		pr, err := server.api.QueryPrognosis(target.Target, target.Data.Period)
		if err != nil {
			log.Printf("api call failed: %v", err)
			return qres
		}

		qres.Datapoints = append(qres.Datapoints, grafana.ResponseTuple{
			Value:     pr.Consumption,
			Timestamp: time.Now().Unix(),
		})
	}

	return qres
}

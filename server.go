package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Server is the http endpoint used by Grafana's SimpleJson plugin
type Server struct {
	api         *Api
	entityCache map[string]string
}

func newServer(api *Api) *Server {
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
	ar := AnnotationsRequest{}
	if err := json.NewDecoder(r.Body).Decode(&ar); err != nil {
		http.Error(w, fmt.Sprintf("json decode failed: %v", err), http.StatusBadRequest)
		return
	}

	resp := []AnnotationResponse{}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("json encode failed: %v", err)
		http.Error(w, fmt.Sprintf("json encode failed: %v", err), http.StatusInternalServerError)
		return
	}
}

func (server *Server) tagKeysHandler(w http.ResponseWriter, r *http.Request) {
	resp := []TagKeyResponse{
		TagKeyResponse{
			Type: "string",
			Text: "group"},
		// TagKeyResponse{
		// 	Type: "string",
		// 	Text: "mode"}
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("json encode failed: %v", err)
		http.Error(w, fmt.Sprintf("json encode failed: %v", err), http.StatusInternalServerError)
		return
	}
}

func (server *Server) tagValuesHandler(w http.ResponseWriter, r *http.Request) {
	resp := []TagValueResponse{
		TagValueResponse{"Current"},
		TagValueResponse{"Consumption"},
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("json encode failed: %v", err)
		http.Error(w, fmt.Sprintf("json encode failed: %v", err), http.StatusInternalServerError)
		return
	}
}

func (server *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	sr := SearchRequest{}
	if err := json.NewDecoder(r.Body).Decode(&sr); err != nil {
		log.Printf("json decode failed: %v", err)
		http.Error(w, fmt.Sprintf("json decode failed: %v", err), http.StatusBadRequest)
		return
	}

	resp := server.executeSearch(sr)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("json encode failed: %v", err)
		http.Error(w, fmt.Sprintf("json encode failed: %v", err), http.StatusInternalServerError)
		return
	}
}

func (server *Server) flattenEntities(result *[]Entity, entities []Entity, parent string) {
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

func (server *Server) populateCache(entities []Entity) {
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

func (server *Server) getPublicEntites() []Entity {
	entities := make([]Entity, 0)
	server.flattenEntities(&entities, server.api.getEntities(), "")
	server.populateCache(entities)
	return entities
}

func (server *Server) executeSearch(sr SearchRequest) []SearchResponse {
	entities := server.getPublicEntites()

	res := []SearchResponse{}
	for _, entity := range entities {
		res = append(res, SearchResponse{
			Text: entity.Title,
			UUID: entity.UUID,
		})
	}

	return res
}

func (server *Server) queryHandler(w http.ResponseWriter, r *http.Request) {
	qr := QueryRequest{}
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
	t := time.Unix(ts/1000, 0)

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

func (server *Server) executeQuery(qr QueryRequest) []QueryResponse {
	res := make([]QueryResponse, len(qr.Targets))
	wg := &sync.WaitGroup{}

	for idx, target := range qr.Targets {
		wg.Add(1)

		go func(idx int, target Target) {
			var context string
			if target.Data.Context != "" {
				context = strings.ToLower(target.Data.Context)
			}

			var qres QueryResponse
			if context == "prognosis" {
				qres = server.queryPrognosis(target)
			} else {
				qres = server.queryData(target, &qr)
			}

			// substitute name
			if text, ok := server.entityCache[qres.Target.(string)]; ok {
				qres.Target = text
			}

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

func (server *Server) queryData(target Target, qr *QueryRequest) QueryResponse {
	qres := QueryResponse{
		Target:     target.Target,
		Datapoints: []ResponseTuple{},
	}

	var group, options string
	if target.Data.Group != "" {
		group = strings.ToLower(target.Data.Group)
	}
	if target.Data.Options != "" {
		options = strings.ToLower(target.Data.Options)
	}

	tuples := server.api.getData(
		target.Target,
		qr.Range.From,
		qr.Range.To,
		group,
		options,
		qr.MaxDataPoints)

	for _, tuple := range tuples {
		if group != "" {
			tuple.Timestamp = roundTimestampMS(tuple.Timestamp, group)
		}

		qres.Datapoints = append(qres.Datapoints, ResponseTuple{
			Timestamp: tuple.Timestamp,
			Value:     tuple.Value,
		})
	}

	return qres
}

func (server *Server) queryPrognosis(target Target) QueryResponse {
	qres := QueryResponse{
		Target:     target.Target,
		Datapoints: []ResponseTuple{},
	}

	if target.Data.Period != "" {
		pr := server.api.getPrognosis(target.Target, target.Data.Period)

		qres.Datapoints = append(qres.Datapoints, ResponseTuple{
			Value:     pr.Consumption,
			Timestamp: time.Now().Unix(),
		})
	}

	return qres
}

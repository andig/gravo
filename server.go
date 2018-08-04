package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

// Server is the http endpoint used by Grafana's SimpleJson plugin
type Server struct {
	api         *Api
	entityCache map[string]string
}

func newServer(api *Api) *Server {
	return &Server{
		api:         api,
		entityCache: make(map[string]string),
	}
}

func (server *Server) rootHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	log.Printf("%v", string(body))
	fmt.Fprintf(w, "ok\n")
}

func (server *Server) annotationsHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	switch r.Method {
	case http.MethodOptions:
	case http.MethodPost:
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
	default:
		http.Error(w, "Bad method; supported OPTIONS, POST", http.StatusBadRequest)
		return
	}

	duration := time.Now().Sub(start)
	log.Printf("%v %v (took %s)", r.Method, r.URL.Path, duration.String())
}

func (server *Server) tagKeysHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	switch r.Method {
	case http.MethodOptions:
	case http.MethodPost:
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
	default:
		http.Error(w, "Bad method; supported OPTIONS, POST", http.StatusBadRequest)
		return
	}

	duration := time.Now().Sub(start)
	log.Printf("%v %v (took %s)", r.Method, r.URL.Path, duration.String())
}

func (server *Server) tagValuesHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	switch r.Method {
	case http.MethodOptions:
	case http.MethodPost:
		resp := []TagValueResponse{
			TagValueResponse{"Current"},
			TagValueResponse{"Consumption"},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("json encode failed: %v", err)
			http.Error(w, fmt.Sprintf("json encode failed: %v", err), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "Bad method; supported OPTIONS, POST", http.StatusBadRequest)
		return
	}

	duration := time.Now().Sub(start)
	log.Printf("%v %v (took %s)", r.Method, r.URL.Path, duration.String())
}

func (server *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	switch r.Method {
	case http.MethodOptions:
	case http.MethodPost:
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
	default:
		http.Error(w, "Bad method; supported OPTIONS, POST", http.StatusBadRequest)
		return
	}

	duration := time.Now().Sub(start)
	log.Printf("%v %v (took %s)", r.Method, r.URL.Path, duration.String())
}

func addEntitiesRecursive(result *[]Entity, entities []Entity, parent string) {
	for _, entity := range entities {
		if entity.Type == "group" {
			addEntitiesRecursive(result, entity.Children, entity.Title)
		} else {
			if parent != "" {
				entity.Title = fmt.Sprintf("%s (%s)", entity.Title, parent)
			}
			*result = append(*result, entity)
		}
	}
}

func (server *Server) executeSearch(sr SearchRequest) []SearchResponse {
	entities := []Entity{}
	addEntitiesRecursive(&entities, server.api.getEntities(), "")

	res := []SearchResponse{}
	for _, entity := range entities {
		// add to cache
		if _, ok := server.entityCache[entity.UUID]; !ok {
			server.entityCache[entity.UUID] = entity.Title
		}

		res = append(res, SearchResponse{
			Text: entity.Title,
			UUID: entity.UUID,
		})
	}

	return res
}

func (server *Server) queryHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	switch r.Method {
	case http.MethodOptions:
	case http.MethodPost:
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
	default:
		http.Error(w, "Bad method; supported OPTIONS, POST", http.StatusBadRequest)
		return
	}

	duration := time.Now().Sub(start)
	log.Printf("%v %v (took %s)", r.Method, r.URL.Path, duration.String())
}

func (server *Server) executeQuery(qr QueryRequest) []QueryResponse {
	res := []QueryResponse{}
	wg := &sync.WaitGroup{}

	for _, target := range qr.Targets {
		wg.Add(1)

		// go server.apiLoader(wg, &res, qr, target)
		go func(wg *sync.WaitGroup, target Target) {
			var group, options string
			data := target.Data
			group, _ = data["group"]
			options, _ = data["options"]
			if options == "" {
				options, _ = data["mode"]
			}

			tuples := server.api.getData(
				target.Target,
				qr.Range.From,
				qr.Range.To,
				group,
				options,
				qr.MaxDataPoints)

			t := target.Target
			if title, ok := server.entityCache[target.Target]; ok {
				t = title
			}

			qtr := &QueryResponse{
				Target:     t,
				Datapoints: []Tuple{},
			}

			for _, tuple := range tuples {
				qtr.Datapoints = append(qtr.Datapoints, Tuple{tuple[1], tuple[0]})
			}

			res = append(res, *qtr)
			wg.Done()
		}(wg, target)
	}

	wg.Wait()
	return res
}

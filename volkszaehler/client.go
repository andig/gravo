package volkszaehler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// Client is the volkszaehler API client
type Client interface {
	Get(endpoint string) (io.ReadCloser, error)
	Post(endpoint string, payload string) (io.ReadCloser, error)
	QueryPublicEntities() ([]Entity, error)
	QueryEntity(entity string) (Entity, error)
	QueryData(uuid string, from time.Time, to time.Time, group string, options string, tuples int) ([]Tuple, error)
	QueryPrognosis(uuid string, period string) (Prognosis, error)
}

type client struct {
	url    string
	client http.Client
	debug  bool
}

// NewClient creates new volkszaehler api client
func NewClient(url string, timeout *time.Duration, debug bool) Client {
	return &client{
		url: url,
		client: http.Client{
			Timeout: *timeout,
		},
		debug: debug,
	}
}

func (api *client) debugResponseBody(resp *http.Response) error {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	log.Print(string(body))
	return nil
}

// Get returns a GET requests body or error. It is the clients responsibility
// to close the response body in case error is not nil
func (api *client) Get(endpoint string) (io.ReadCloser, error) {
	url := api.url + endpoint

	start := time.Now()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, err
	}

	duration := time.Since(start)
	log.Printf("GET %s (%dms)", url, duration.Nanoseconds()/1e6)

	if api.debug {
		if err := api.debugResponseBody(resp); err != nil {
			return nil, err
		}
	}

	return resp.Body, nil
}

// Post returns a GET requests body or error. It is the clients responsibility
// to close the response body in case error is not nil
func (api *client) Post(endpoint string, payload string) (io.ReadCloser, error) {
	url := api.url + endpoint

	req, err := http.NewRequest("POST", url, strings.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-type", "application/json")
	req.Header.Add("Accept", "application/json")

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, err
	}

	if api.debug {
		if err := api.debugResponseBody(resp); err != nil {
			return nil, err
		}
	}

	return resp.Body, nil
}

// QueryPublicEntities retrieves public entities from middleware
func (api *client) QueryPublicEntities() ([]Entity, error) {
	body, err := api.Get("/entity.json")
	if err != nil {
		return []Entity{}, err
	}
	defer func() {
		_ = body.Close() // close body after checking for error
	}()

	er := EntitiesResponse{}
	if err := json.NewDecoder(body).Decode(&er); err != nil {
		log.Printf("json decode failed: %v", err)
		return []Entity{}, err
	}

	if er.Exception.Message != "" {
		return []Entity{}, errors.New("api exception: " + er.Exception.Message)
	}

	return er.Entities, nil
}

// QueryEntity retrieves entitiy by uuid
func (api *client) QueryEntity(entity string) (Entity, error) {
	context := fmt.Sprintf("/entity/%s.json", entity)
	body, err := api.Get(context)
	if err != nil {
		return Entity{}, err
	}
	defer func() {
		_ = body.Close() // close body after checking for error
	}()

	er := EntityResponse{}
	if err := json.NewDecoder(body).Decode(&er); err != nil {
		return Entity{}, err
	}

	if er.Exception.Message != "" {
		return Entity{}, errors.New("api exception: " + er.Exception.Message)
	}

	return er.Entity, nil
}

// QueryData retrieves data for specified timeframe and parameters
func (api *client) QueryData(uuid string, from time.Time, to time.Time,
	group string, options string, tuples int,
) ([]Tuple, error) {
	f := from.Unix()
	t := to.Unix()
	url := fmt.Sprintf("/data/%s.json?from=%d&to=%d", uuid, f*1000, t*1000)

	if tuples > 0 {
		url += fmt.Sprintf("&tuples=%d", tuples)
	}

	if group != "" {
		url += "&group=" + group
	}

	if options != "" {
		url += "&options=" + options
	}

	body, err := api.Get(url)
	if err != nil {
		return []Tuple{}, err
	}
	defer func() {
		_ = body.Close() // close body after checking for error
	}()

	dr := DataResponse{}
	if err := json.NewDecoder(body).Decode(&dr); err != nil {
		return []Tuple{}, err
	}

	if dr.Exception.Message != "" {
		return []Tuple{}, errors.New("api exception: " + dr.Exception.Message)
	}

	return dr.Data.Tuples, nil
}

// QueryPrognosis retrieves prognosis from middleware
func (api *client) QueryPrognosis(uuid string, period string) (Prognosis, error) {
	url := fmt.Sprintf("/prognosis/%s.json?period=%s", uuid, period)

	body, err := api.Get(url)
	if err != nil {
		return Prognosis{}, err
	}
	defer func() {
		_ = body.Close() // close body after checking for error
	}()

	pr := PrognosisResponse{}
	if err := json.NewDecoder(body).Decode(&pr); err != nil {
		return Prognosis{}, err
	}

	if pr.Exception.Message != "" {
		return Prognosis{}, errors.New("api exception: " + pr.Exception.Message)
	}

	return pr.Prognosis, nil
}

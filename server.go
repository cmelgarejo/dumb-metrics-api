package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	srvAddr := getStringENV("HOST", "") + ":" + getStringENV("PORT", "8080")
	mdb := make(MetricsDB, 0)
	http.HandleFunc("/", metricsHandler(&mdb, getIntENV("DATA_TIMEOUT_MINUTES", 60)))
	log.Printf("Running metrics server at [%s]\n", srvAddr)
	log.Fatalf("FATAL: %q\n", http.ListenAndServe(srvAddr, nil))
}

// Metric struct

// Metric represents a metric
type Metric struct {
	Value     int       `json:"value"`
	Timestamp time.Time `json:"-"`
}

// MetricsDB would be out place-in db struct
type MetricsDB map[string][]Metric

// SumMetric returns an metric instance with the sum value of a specific key
func (mdb MetricsDB) SumMetric(key string, to int) *Metric {
	mdb.checkDataTimeout(key, to)
	metric := Metric{}
	for _, m := range mdb[key] {
		metric.Value += m.Value
	}
	return &metric
}

func (mdb MetricsDB) checkDataTimeout(key string, to int) {
	remIds := make([]int, 0)
	for i, m := range mdb[key] {
		// If a metric times out
		if int(time.Now().Sub(m.Timestamp).Minutes()) > to {
			// Add to the "remove Id" slice
			remIds = append(remIds, i)
		}
	}
	if len(remIds) > 0 {
		// since order is not important, just wipe the object in that place and substract a place in the slice/array
		// this is less costly then traversing the whole slice/array and ask if the element timed out and copy into a new one.
		for _, i := range remIds {
			mdb[key][i] = mdb[key][len(mdb[key])-1]
			mdb[key] = mdb[key][:len(mdb[key])-1]
		}
	}
}

// HTTP Hanlders

var (
	mx sync.Mutex
)

func metricsHandler(mdb *MetricsDB, to int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[server] got a request: %s - %s\n", r.Method, r.URL.Path)
		if !strings.HasPrefix(r.URL.Path, "/metric/") {
			http.NotFound(w, r)
			return
		}
		switch r.Method {
		case http.MethodGet:
			getMetrics(mdb, to, w, r)
		case http.MethodPost:
			postMetrics(mdb, w, r)
		}
	}
}

func getMetrics(mdb *MetricsDB, to int, w http.ResponseWriter, r *http.Request) {
	if path.Base(r.URL.Path) == "sum" {
		metricKey := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/metric/"), "/sum")
		mx.Lock() // lock the access to our "db"
		defer mx.Unlock()
		if _, ok := (*mdb)[metricKey]; ok { // If the metric exists
			b, err := json.Marshal(mdb.SumMetric(metricKey, to)) // get the sum, also clean data
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			w.Header().Add("Content-Type", "application/json")
			fmt.Fprintf(w, string(b))
			return
		}
	}
	http.NotFound(w, r)
}

func postMetrics(mdb *MetricsDB, w http.ResponseWriter, r *http.Request) {
	var m Metric
	key := path.Base(r.URL.Path)
	if strings.TrimPrefix(r.URL.Path, "/metric/") != key {
		http.Error(w, "Bad request", http.StatusBadRequest)
	}
	if r.Header.Get("Content-Type") != "" {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Content-Type header is not application/json", http.StatusUnsupportedMediaType)
			return
		}
	}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	m.Timestamp = time.Now()
	mx.Lock() // lock the access to our "db"
	defer mx.Unlock()
	if _, ok := (*mdb)[key]; !ok {
		(*mdb)[key] = make([]Metric, 0)
	}
	(*mdb)[key] = append((*mdb)[key], m)
	// Should write header with CREATED 201, not 200, this is just a note on REST API's.
	// w.WriteHeader(http.StatusCreated)
	w.Header().Add("Content-Type", "application/json")
	fmt.Fprintf(w, "{}")
}

// Program helpers

func getStringENV(key string, def string) string {
	if os.Getenv(key) == "" {
		return def
	}
	return os.Getenv(key)
}

func getIntENV(key string, def int) int {
	v, err := strconv.Atoi(os.Getenv(key))
	if os.Getenv(key) == "" || err != nil {
		return def
	}
	return v
}

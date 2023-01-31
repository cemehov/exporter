package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Total struct {
	Value    int    `json:"value"`
	Relation string `json:"relation"`
}

type Hits struct {
	Total    Total    `json:"total"`
	MaxScore string   `json:"max_score"`
	Hits     []string `json:"hits"`
}

type Shards struct {
	Total      int `json:"total"`
	Successful int `json:"successful"`
	Skipped    int `json:"skipped"`
	Failed     int `json:"failed"`
}

type Resp struct {
	Took     int    `json:"took"`
	TimedOut bool   `json:"timed_out"`
	Hits     Hits   `json:"hits"`
	Shards   Shards `json:"_shards"`
}

var jsonStr = []byte(`{
        "from" : 0,
        "size" : 0,
        "query" : {
          "bool": {
            "must": [
               {
               "match_phrase" : {
                 "message": "Notification has been send OK with json"
                 }
               },
               {
                "range": {
                  "@timestamp": {
                    "time_zone": "+03:00",
                    "gte" : "now-5m",
                    "lt" : "now"
                  }
                }
              }
            ]
          }
        }
      }
`)

func getRequestsOivs() int {

	req, err := http.NewRequest("POST", "http://localhost:9200/logstash*/_search", bytes.NewBuffer(jsonStr))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	// fmt.Println(resp)
	// fmt.Println("response Status:", resp.Status)
	// fmt.Println("response Headers:", resp.Header)

	body, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println("response Body:", string(body))

	var respns Resp
	err = json.Unmarshal(body, &respns)
	if err != nil {
		panic(err)
	}

	// fmt.Println(respns.Hits.Total.Value)

	return respns.Hits.Total.Value

}

func recordMetrics() {
	go func() {
		for {
			opsProcessed.Set(float64(getRequestsOivs()))
			time.Sleep(300 * time.Second)
		}
	}()
}

var (
	opsProcessed = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "requests_oivs",
	})
)

func main() {

	recordMetrics()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)

}

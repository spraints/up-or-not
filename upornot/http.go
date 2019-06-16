package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
)

func buildHTTPHandler(models []*model) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/api/status", apiStatus(models[0]))
	mux.Handle("/api/targets", apiTargets(models))
	for _, m := range models {
		prefix := apiTargetPrefix(m)
		mux.Handle(prefix, apiStatus(m))
		mux.Handle(prefix+"/recent", apiRecents(m))
	}
	// ? mux.HandleFunx("", homePage(m))
	return mux
}

func apiTargetPrefix(m *model) string {
	return "/api/target/" + url.PathEscape(m.TargetIP)
}

func apiTargets(models []*model) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var res struct {
			TargetIPs []string `json:"targets"`
		}

		for _, m := range models {
			res.TargetIPs = append(res.TargetIPs, apiTargetPrefix(m))
		}

		apiRes(w, res)
	})
}

func apiStatus(m *model) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var res struct {
			Count     int     `json:"count"`
			AvgMillis float64 `json:"avg_ms"`
			OKCount   int     `json:"ok"`
			Buckets   [5]struct {
				MaxMillis int `json:"max_ms,omitempty"`
				Count     int `json:"count"`
			} `json:"buckets"`
		}
		res.Buckets[0].MaxMillis = 60
		res.Buckets[1].MaxMillis = 100
		res.Buckets[2].MaxMillis = 200
		res.Buckets[3].MaxMillis = 1000

		vals := m.Get()
		res.Count = len(vals)
		var totalMs float64
		for _, val := range vals {
			if val.Result == ok {
				res.OKCount++
				ms := (val.Duration.Seconds() * 1000.0)
				totalMs += ms
				for i := 0; i < len(res.Buckets); i++ {
					max := res.Buckets[i].MaxMillis
					if max == 0 || ms < float64(max) {
						res.Buckets[i].Count++
						break
					}
				}
			}
		}
		if res.OKCount > 0 {
			res.AvgMillis = totalMs / float64(res.OKCount)
		}

		apiRes(w, res)
	})
}

func apiRecents(m *model) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var res struct {
			Recent []dataPoint `json:"recent"`
		}

		res.Recent = m.Get()

		apiRes(w, res)
	})
}

func apiRes(w http.ResponseWriter, res interface{}) {
	if json, err := json.Marshal(res); err != nil {
		logerr(err, "writing api response")
		http.Error(w, "response marshaling failed", http.StatusInternalServerError)
	} else {
		w.Header().Add("Content-Type", "application/json")
		w.Write(json)
	}
}

func serveHTTP(ctx context.Context, server *http.Server) error {
	log.Printf("HTTP server started (%s)", server.Addr)
	defer log.Printf("HTTP server stopped (%s)", server.Addr)

	done := make(chan error)
	go func() {
		defer close(done)
		done <- server.ListenAndServe()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		log.Printf("Shutdown HTTP server")
		return server.Shutdown(ctx)
	}
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"golang.org/x/sync/errgroup"
	"net/http"
)

type Server struct {
	ctx    context.Context
	cancel context.CancelFunc

	dataCh chan string
	eg     *errgroup.Group
}

func main() {
	srv := &Server{
		dataCh: make(chan string, 20),
		eg:     &errgroup.Group{},
	}
	defer close(srv.dataCh)

	srv.ctx, srv.cancel = context.WithCancel(context.Background())
	defer srv.cancel()

	r := mux.NewRouter()
	r.HandleFunc("/addWorkers", srv.addWorkersHandler).Methods(http.MethodPost)
	r.HandleFunc("/addData", srv.addDataHandler).Methods(http.MethodPost)

	fmt.Println("Server listening on http://localhost:8080")
	http.ListenAndServe(":8080", r)
}

//	{
//	   "num": 3
//	}
func (s *Server) addWorkersHandler(w http.ResponseWriter, r *http.Request) {
	var num struct {
		Num int `json:"num"`
	}
	if err := json.NewDecoder(r.Body).Decode(&num); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for i := 0; i < num.Num; i++ {
		s.eg.Go(func() error {
			return Worker(s.ctx, s.dataCh, fmt.Sprintf("number %d", i))
		})
	}
}

//	{
//	   "data": [
//	       "wow",
//	       "bless"
//	   ]
//	}
func (s *Server) addDataHandler(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Data []string `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, v := range data.Data {
		s.dataCh <- v
	}
}

func Worker(ctx context.Context, inCh <-chan string, name string) error {
	fmt.Printf("Worker %s started\n", name)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case v := <-inCh:
			fmt.Printf("Worker %s received %s\n", name, v)
		}
	}
}

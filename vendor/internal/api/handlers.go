package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"strconv"
)

func handleErrors(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func (s *Server) List() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		cms, err := s.DB.List(ctx)
		if err != nil {
			handleErrors(w, err)
			return
		}

		json, err := json.Marshal(&cms)
		if err != nil {
			handleErrors(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(json)
	}
}

func (s *Server) Get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		vars := mux.Vars(r)

		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "bad 'id' url argument", http.StatusBadRequest)
			return
		}

		cm, err := s.DB.Get(ctx, IDType(id))
		if err != nil {
			handleErrors(w, err)
			return
		}
		json, err := json.Marshal(&cm)
		if err != nil {
			handleErrors(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(json)
	}
}

func (s *Server) Create() http.HandlerFunc {
	type request struct {
		Brand   string `json:"brand"`
		Model   string `json:"model"`
		Price   int64  `json:"price"`
		Status  Status `json:"status"`
		Mileage int64  `json:"mileage"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			handleErrors(w, err)
			return
		}

		var req request
		err = json.Unmarshal(bytes, &req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		cm, err := s.DB.Create(ctx, req.Brand, req.Model, req.Price, req.Status, req.Mileage)
		if err != nil {
			handleErrors(w, err)
			return
		}

		json, err := json.Marshal(cm)
		if err != nil {
			handleErrors(w, err)
			return
		}

		w.WriteHeader(http.StatusAccepted)
		w.Write(json)
	}
}

func (s *Server) Update() http.HandlerFunc {
	type request struct {
		Brand   string `json:"brand"`
		Model   string `json:"model"`
		Price   int64  `json:"price"`
		Status  Status `json:"status"`
		Mileage int64  `json:"mileage"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			handleErrors(w, err)
			return
		}

		var req request
		err = json.Unmarshal(bytes, &req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "bad 'id' url argument", http.StatusBadRequest)
			return
		}

		cm, err := s.DB.Update(ctx, IDType(id), req.Brand, req.Brand, req.Price, req.Status, req.Mileage)
		if err != nil {
			handleErrors(w, err)
			return
		}

		json, err := json.Marshal(cm)
		if err != nil {
			handleErrors(w, err)
			return
		}

		w.WriteHeader(http.StatusAccepted)
		w.Write(json)
	}
}

func (s *Server) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "bad 'id' url argument", http.StatusBadRequest)
			return
		}

		err = s.DB.Delete(ctx, IDType(id))
		if err != nil {
			handleErrors(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
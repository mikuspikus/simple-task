package api

func (s *Server) routes() {
	router := s.Router.PathPrefix("/tt/v0").Subrouter()

	router.HandleFunc("/cars", s.List()).Methods("GET")
	router.HandleFunc("/cars/{id:[0-9]+}", s.Get()).Methods("GET")
	router.HandleFunc("/cars", s.Create()).Methods("POST")
	router.HandleFunc("/cars/{id:[0-9]+}", s.Update()).Methods("PUT")
	router.HandleFunc("/cars/{id:[0-9]+}", s.Delete()).Methods("DELETE")
}

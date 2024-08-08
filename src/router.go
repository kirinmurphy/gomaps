package main

import (
	"encoding/json"
	"gomap/src/locationHelpers"
	"log"
	"net/http"
	"text/template"

	"github.com/gorilla/mux"
)

func initRouter(locationStore *LocationStore, setLocations func([]locationHelpers.Location)) {
	r := mux.NewRouter()

	fs := http.FileServer(http.Dir("src/templates/dist"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))
	http.Handle("/output.css", http.StripPrefix("/", fs))

	r.HandleFunc("/", homeHandler)

	r.HandleFunc("/locations", func(w http.ResponseWriter, r *http.Request) {
		locationsHandler(w, r, locationStore)
	})

	r.HandleFunc("/reload-locations", func(w http.ResponseWriter, r *http.Request) {
		reloadLocationsHandler(w, r, setLocations, locationStore)
	}).Methods("GET")

	r.PathPrefix("/src/templates/").Handler(http.StripPrefix("/src/templates/", http.FileServer(http.Dir("src/templates"))))

	log.Println("Server starting on port 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("src/templates/dist/index.html"))
	tmpl.Execute(w, nil)
}

func locationsHandler(w http.ResponseWriter, r *http.Request, locationStore *LocationStore) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(locationStore.GetLocations())
}

func reloadLocationsHandler(w http.ResponseWriter, r *http.Request, setLocations func([]locationHelpers.Location), locationStore *LocationStore) {
	reloadedLocations, err := locationHelpers.LoadLocations()
	if err != nil {
		log.Printf("Failed to reload locations: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	currentLocations := locationStore.GetLocations()
	diff := locationHelpers.DiffLocations(currentLocations, reloadedLocations)

	setLocations(reloadedLocations)

	log.Printf("current locations: %d", len(currentLocations))
	log.Printf("Reloaded %d locations", len(reloadedLocations))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(diff)
}

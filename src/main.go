package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"tendermint_proposal_monitor/config"
	"tendermint_proposal_monitor/monitor"
	"tendermint_proposal_monitor/proposals"
	"tendermint_proposal_monitor/services"
)

var (
	cfg       *config.Configurations
	globalErr error
	useMock   bool
)

func init() {
	flag.BoolVar(&useMock, "mock", false, "Use mock data for testing")
	configFile := flag.String("config", getEnv("CONFIG_FILE", "src/config/config.yml"), "Path to configuration file")
	flag.Parse()

	log.Println("Starting Proposal Monitor Service...")
	log.Printf("Using configuration file: %s\n", *configFile)

	cfg, globalErr = config.LoadConfig(*configFile)
	if globalErr != nil {
		log.Fatalf("Error loading config: %v", globalErr)
	}

	log.Printf("Configuration loaded successfully.")
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func triggerMonitor(w http.ResponseWriter, r *http.Request) {
	if globalErr != nil {
		http.Error(w, "Configuration error, unable to run monitor", http.StatusInternalServerError)
		return
	}

	firestoreHandler, err := proposals.New(cfg)
	if err != nil {
		log.Printf("Error creating FirestoreHandler: %v", err)
		http.Error(w, "Error creating FirestoreHandler", http.StatusInternalServerError)
		return
	}
	s := services.New(firestoreHandler, cfg)
	h := monitor.NewHandler(s)

	// Check for mock query parameter
	mock := r.URL.Query().Get("mock")
	useMock := mock == "true"

	err = h.Run(cfg, useMock)
	if err != nil {
		log.Printf("Error running monitor: %v", err)
		http.Error(w, "Error running monitor", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Monitor triggered successfully"))
}

func healthcheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	http.HandleFunc("/trigger-monitor", triggerMonitor)
	http.HandleFunc("/health", healthcheck)
	log.Println("Server started on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

package main

import (
	"flag"
	"log"
	"tendermint_proposal_monitor/config"
	"tendermint_proposal_monitor/monitor"
)

func main() {
	useMock := flag.Bool("mock", false, "Use mock data for testing")
	configFile := flag.String("config", "config/config.yml", "Path to configuration file")
	flag.Parse()

	log.Println("Starting Proposal Monitor Service...")
	log.Printf("Using configuration file: %s\n", *configFile)

	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	log.Printf("Configuration loaded successfully. Check interval: %d seconds\n", cfg.CheckInterval)

	err = monitor.Run(cfg, *useMock)
	if err != nil {
		log.Fatalf("Error running monitor: %v", err)
	}

	log.Println("Proposal Monitor Service is running.")
}

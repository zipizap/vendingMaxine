package main

import "time"

type StatusBlock struct {
	Name                   string    `yaml:"Name"`
	StartTime              time.Time `yaml:"StartTime"`
	LatestUpdateTime       time.Time `yaml:"LatestUpdateTime"`
	LatestUpdateStatus     string    `yaml:"LatestUpdateStatus"`
	LatestUpdateStatusInfo string    `yaml:"LatestUpdateStatusInfo"`
	LatestUpdateUml        string    `yaml:"LatestUpdateUml"`
	LatestUpdateData       struct{}  `yaml:"LatestUpdateData"`
}

// "encoding/json"
type UserviceFlow struct {
	Kind   string `yaml:"Kind"`
	Name   string `yaml:"Name"`
	Status struct {
		Overall           StatusBlock   `yaml:"Overall"`
		ProcessingEngines []StatusBlock `yaml:"ProcessingEngines"`
	} `yaml:"Status"`
}

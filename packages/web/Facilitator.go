package web

import (
	"vendingMaxine/packages/collection"

	"go.uber.org/zap"
)

type Facilitator struct{}

func NewFacilitator() (*Facilitator, error) {
	return &Facilitator{}, nil
}

func (f *Facilitator) InitSetup(dbFilepath string, catalogDefaultName string, catalogDefaultDirpath string, zapSugaredLogger *zap.SugaredLogger) {
	// collection.Facilitator: Create and InitSetup
	slog = zapSugaredLogger
	cf, _ := collection.NewFacilitator()
	cf.InitSetup(dbFilepath, catalogDefaultName, catalogDefaultDirpath, zapSugaredLogger)
}

// Starts http web+api server (blocks forever)
//
//	go f.StartServer(":8080")
func (f *Facilitator) StartServer(listeningAddress string) {
	startServer(listeningAddress)
}

// Stop http web+api server (gracefully within 10s, or ungracefully after 10s)
//
//	err := f.StopServer()
func (f *Facilitator) StopServer() error {
	return closeServer()
}

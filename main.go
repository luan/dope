package main

import (
	"runtime"

	"github.com/luan/idope/config_finder"
	"github.com/luan/idope/fetcher"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	bbsClient, err := config_finder.NewBBS()
	if err != nil {
		panic(err)
	}

	noaaClient, err := config_finder.NewNoaaConsumer()
	if err != nil {
		panic(err)
	}

	ui := NewUI()
	ui.Setup()
	defer ui.Close()

	go func() {
		defer ui.Close()
		fetcher := fetcher.NewFetcher(bbsClient, noaaClient)
		// for {
		state, err := fetcher.Fetch()
		if err != nil {
			panic(err)
		}
		ui.SetState(&state)
		// time.Sleep(time.Second)
		// }
	}()

	ui.Loop()
}

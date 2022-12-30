package main

import (
	"fmt"
	"log"
	"time"

	aw "github.com/deanishe/awgo"
)

func run(wf *aw.Workflow) {
	args := wf.Args()
	log.Printf("args: %s", wf.Args())

	if len(args) > 0 && args[0] != "" {
		query := args[0]
		wf.Filter(query)
	}

	if wf.IsEmpty() {
		wf.WarnEmpty("No matching found", "")
		return
	}

	// Send results to Alfred
	wf.SendFeedback()
}

func main() {
	var (
		wf  *aw.Workflow
		cfg *aw.Config
	)

	reload := func() ([]byte, error) {
		return []byte{}, nil
	}
	wf.Cache.LoadOrStore("cachedb.kdbx", time.Second*5, reload)

	wf = aw.New()
	cfg = aw.NewConfig()
	fmt.Println(cfg)

	wf.Run(func() {
		run(wf)
	})
}

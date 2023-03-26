package main

import (
	"log"
	"time"

	aw "github.com/deanishe/awgo"
)

var (
	wf  *aw.Workflow
	cfg *aw.Config
)

func init() {
	wf = aw.New()
	cfg = aw.NewConfig()
}

func run(wf *aw.Workflow) {
	args := wf.Args()
	log.Printf("args: %s", wf.Args())

	if len(args) > 0 && args[0] != "" {
		query := args[0]
		wf.Filter(query)
	}

	db, err := HTTPGetFile()
	if err != nil {
		wf.FatalError(err)
		return
	}
	log.Print(db)

	if wf.IsEmpty() {
		wf.WarnEmpty("No matching found", "")
		return
	}

	// Send results to Alfred
	wf.SendFeedback()
}

func main() {
	reload := func() ([]byte, error) {
		return []byte{}, nil
	}
	wf.Cache.LoadOrStore("cachedb.kdbx", time.Second*5, reload)

	wf.Run(func() {
		run(wf)
	})
}

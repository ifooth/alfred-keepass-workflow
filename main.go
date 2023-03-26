package main

import (
	"log"

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

	kee := NewKee(GetKeepassURL(cfg), GetKesspassPwd(cfg))
	kee.Load(wf)

	if wf.IsEmpty() {
		wf.WarnEmpty("No matching found", "")
		return
	}

	// Send results to Alfred
	wf.SendFeedback()
}

func main() {
	wf.Run(func() {
		run(wf)
	})
}

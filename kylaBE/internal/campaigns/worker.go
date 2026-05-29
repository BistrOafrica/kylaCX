package campaigns

import (
	"fmt"
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// WorkerHandle wraps a started Temporal worker so main.go can stop it on
// shutdown.
type WorkerHandle struct {
	w worker.Worker
}

func (h *WorkerHandle) Stop() {
	if h == nil || h.w == nil {
		return
	}
	h.w.Stop()
}

// StartWorker registers CampaignExecutionWorkflow + all campaign activities on
// the supplied task queue and starts the worker in the background.
//
// Temporal allows multiple workers on the same task queue; we deliberately use
// a separate worker from the automation engine so a slow audience resolution
// can't starve the workflow worker pool. They share kyla-automation by default
// but run in independent goroutines.
//
// Returns (nil, nil) when temporalClient is nil — matches the graceful
// degradation pattern from automation.StartWorker.
func StartWorker(temporalClient client.Client, taskQueue string, deps ActivityDeps) (*WorkerHandle, error) {
	if temporalClient == nil {
		log.Println("campaigns worker: temporal client nil; worker not started")
		return nil, nil
	}
	if taskQueue == "" {
		return nil, fmt.Errorf("campaigns worker: task queue is required")
	}

	w := worker.New(temporalClient, taskQueue, worker.Options{})

	// Workflows
	w.RegisterWorkflow(CampaignExecutionWorkflow)

	// Activities — registered as struct receivers so Temporal uses the method
	// name (e.g. "ResolveAudience") as the activity name.
	w.RegisterActivity(&ResolveAudienceActivity{Deps: deps})
	w.RegisterActivity(&LoadSendBatchActivity{Deps: deps})
	w.RegisterActivity(&SendRecipientActivity{Deps: deps})
	w.RegisterActivity(&FinaliseCampaignActivity{Deps: deps})

	if err := w.Start(); err != nil {
		return nil, fmt.Errorf("campaigns worker start: %w", err)
	}
	log.Printf("campaigns worker: started on task queue %q", taskQueue)
	return &WorkerHandle{w: w}, nil
}

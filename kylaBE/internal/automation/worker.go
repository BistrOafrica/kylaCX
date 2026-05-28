package automation

import (
	"fmt"
	"log"

	"kyla-be/internal/automation/activities"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// WorkerHandle wraps a started Temporal worker so the caller can stop it on
// shutdown. The underlying worker is started in a background goroutine when
// StartWorker returns successfully.
type WorkerHandle struct {
	w worker.Worker
}

// Stop gracefully shuts the worker down. Safe to call multiple times.
func (h *WorkerHandle) Stop() {
	if h == nil || h.w == nil {
		return
	}
	h.w.Stop()
}

// StartWorker registers the AutomationWorkflow plus all activities on the
// configured task queue and starts the worker in the background.
//
// Returns (nil, nil) when temporalClient is nil — this matches the graceful
// degradation pattern used for NATS in main.go: the binary still boots when
// Temporal is unavailable, the workflow execution surface just goes dark.
func StartWorker(temporalClient client.Client, taskQueue string, deps activities.Deps) (*WorkerHandle, error) {
	if temporalClient == nil {
		log.Println("automation worker: temporal client nil; worker not started")
		return nil, nil
	}
	if taskQueue == "" {
		return nil, fmt.Errorf("automation worker: task queue is required")
	}

	w := worker.New(temporalClient, taskQueue, worker.Options{})

	// Workflows
	w.RegisterWorkflow(AutomationWorkflow)

	// Activities — each registered as a struct so Temporal uses the method
	// name (e.g. "UpdateObject") as the activity name.
	w.RegisterActivity(&activities.UpdateObjectActivity{Deps: deps})
	w.RegisterActivity(&activities.AssignUserActivity{Deps: deps})
	w.RegisterActivity(&activities.CreateObjectActivity{Deps: deps})
	w.RegisterActivity(&activities.CreateTaskActivity{Deps: deps})
	w.RegisterActivity(&activities.SendMessageActivity{Deps: deps})
	w.RegisterActivity(&activities.InvokeWebhookActivity{Deps: deps})
	w.RegisterActivity(&activities.SetSLAActivity{Deps: deps})
	w.RegisterActivity(&activities.SendNotificationActivity{Deps: deps})
	w.RegisterActivity(&activities.RunAISkillActivity{Deps: deps})

	if err := w.Start(); err != nil {
		return nil, fmt.Errorf("automation worker start: %w", err)
	}
	log.Printf("automation worker: started on task queue %q", taskQueue)
	return &WorkerHandle{w: w}, nil
}

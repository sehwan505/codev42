package handler

import (
	"context"
	"fmt"
	"time"

	"codev42-implementation/configs"
	"codev42-implementation/pb"
	"codev42-implementation/queue"
	"codev42-implementation/service"
)

type ImplementationHandler struct {
	pb.UnimplementedImplementationServiceServer
	Config      configs.Config
	workerAgent *service.WorkerAgent
	jobQueue    *queue.JobQueue
}

func NewImplementationHandler(config configs.Config) *ImplementationHandler {
	workerAgent := service.NewWorkerAgent(config.OpenAiKey)
	jobQueue := queue.NewJobQueue()

	return &ImplementationHandler{
		Config:      config,
		workerAgent: workerAgent,
		jobQueue:    jobQueue,
	}
}

// ImplementPlan starts async implementation and returns job ID
func (h *ImplementationHandler) ImplementPlan(ctx context.Context, req *pb.ImplementPlanRequest) (*pb.ImplementPlanResponse, error) {
	// Create a new job
	job := h.jobQueue.CreateJob(req.DevPlanId)

	// Start async processing
	go h.processImplementation(job.ID, req.DevPlanId)

	return &pb.ImplementPlanResponse{
		JobId:   job.ID,
		Status:  string(queue.JobStatusPending),
		Message: "Implementation job started",
	}, nil
}

// GetImplementationStatus returns the current status of a job
func (h *ImplementationHandler) GetImplementationStatus(ctx context.Context, req *pb.GetImplementationStatusRequest) (*pb.GetImplementationStatusResponse, error) {
	job, err := h.jobQueue.GetJob(req.JobId)
	if err != nil {
		return nil, fmt.Errorf("failed to get job status: %v", err)
	}

	return &pb.GetImplementationStatusResponse{
		JobId:       job.ID,
		Status:      string(job.Status),
		Progress:    job.Progress,
		CurrentStep: job.CurrentStep,
		CreatedAt:   job.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   job.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// GetImplementationResult returns the result of a completed job
func (h *ImplementationHandler) GetImplementationResult(ctx context.Context, req *pb.GetImplementationResultRequest) (*pb.GetImplementationResultResponse, error) {
	job, err := h.jobQueue.GetJob(req.JobId)
	if err != nil {
		return nil, fmt.Errorf("failed to get job result: %v", err)
	}

	response := &pb.GetImplementationResultResponse{
		JobId:  job.ID,
		Status: string(job.Status),
		Error:  job.Error,
	}

	if job.CompletedAt != nil {
		response.CompletedAt = job.CompletedAt.Format(time.RFC3339)
	}

	if job.Result != nil {
		response.Code = job.Result.Code

		// Convert diagrams
		for _, diagram := range job.Result.Diagrams {
			response.Diagrams = append(response.Diagrams, &pb.Diagram{
				Diagram: diagram.Diagram,
				Type:    diagram.Type,
			})
		}

		// Convert explained segments
		for _, segment := range job.Result.ExplainedSegments {
			response.ExplainedSegments = append(response.ExplainedSegments, &pb.ExplainedSegment{
				StartLine:   segment.StartLine,
				EndLine:     segment.EndLine,
				Explanation: segment.Explanation,
			})
		}
	}

	return response, nil
}

// processImplementation is the async worker that processes the implementation
func (h *ImplementationHandler) processImplementation(jobID string, devPlanID int64) {
	// Update job status to processing
	h.jobQueue.UpdateJob(jobID, queue.JobStatusProcessing, 10, "Fetching development plan")

	// TODO: Call Plan Service to fetch plan data by devPlanID
	// For now, create a dummy plan for testing
	plans := []service.Plan{
		{
			ClassName: "Example",
			Annotations: []service.Annotation{
				{
					Name:        "exampleMethod",
					Description: "An example method",
					Params:      "string input",
					Returns:     "string",
				},
			},
		},
	}

	h.jobQueue.UpdateJob(jobID, queue.JobStatusProcessing, 30, "Implementing code")

	// Call worker agent to implement
	results, err := h.workerAgent.ImplementPlan("go", plans)
	if err != nil {
		h.jobQueue.SetJobError(jobID, err)
		return
	}

	h.jobQueue.UpdateJob(jobID, queue.JobStatusProcessing, 70, "Combining implementations")

	// Combine results (simplified - just take first result for now)
	var code string
	if len(results) > 0 && results[0] != nil {
		code = results[0].Code
	}

	h.jobQueue.UpdateJob(jobID, queue.JobStatusProcessing, 90, "Finalizing")

	// Set job result
	result := &queue.JobResult{
		Code:              code,
		Diagrams:          []queue.Diagram{},          // TODO: Generate diagrams
		ExplainedSegments: []queue.ExplainedSegment{}, // TODO: Analyze code segments
	}

	h.jobQueue.SetJobResult(jobID, result)
	h.jobQueue.UpdateJob(jobID, queue.JobStatusCompleted, 100, "Completed")
}

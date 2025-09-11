package runcomfyserverless

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

const (
	timeFormat = "2006-01-02T15:04:05.999999-07:00"
)

/*
	Inference
*/

// Override defines a single override to send to the deployed workflow
type Override struct {
	NodeID string
	Inputs map[string]any
}

type overridePayload struct {
	Overrides map[string]map[string]map[string]any `json:"overrides"`
}

type inferenceRespPayload struct {
	RequestID string `json:"request_id"`
	StatusURL string `json:"status_url"`
	ResultURL string `json:"result_url"`
	CancelURL string `json:"cancel_url"`
	APIError
}

// Start allows to submit a request to a deployment
func (d *Deployment) Start(ctx context.Context, overrides []Override) (requestID string, err error) {
	// build payload
	payload := overridePayload{
		Overrides: make(map[string]map[string]map[string]any, len(overrides)),
	}
	for _, override := range overrides {
		payload.Overrides[override.NodeID] = map[string]map[string]any{
			"inputs": override.Inputs,
		}
	}
	// execute request
	var resp inferenceRespPayload
	if err = d.request(ctx, "POST", "inference", payload, &resp); err != nil {
		err = fmt.Errorf("failed to perform the HTTP request: %w", err)
		return
	}
	// extract result
	if resp.Code == 0 {
		requestID = resp.RequestID
	} else {
		err = resp.APIError
	}
	return
}

/*
	Status
*/

// RequestStatus represents the status of a serverless request initiated with Start()
type RequestStatus string

const (
	RequestStatusInQueue    RequestStatus = "in_queue"
	RequestStatusInProgress RequestStatus = "in_progress"
	RequestStatusCompleted  RequestStatus = "completed" // only for Status(), on Result() it is replaced by RequestStatusSucceeded or RequestStatusFailed
	RequestStatusCanceled   RequestStatus = "canceled"
)

// StatusResponse contains the status of a request and additional information about its position in the queue.
type StatusResponse struct {
	Status        RequestStatus `json:"status"`
	QueuePosition int           `json:"queue_position"`
}

type statusResponsePayload struct {
	StatusResponse
	RequestID string `json:"request_id"`
	ResultURL string `json:"result_url"`
	StatusURL string `json:"status_url"`
	APIError
}

// Status retrieves the current status of a request.
func (d *Deployment) Status(ctx context.Context, requestID string) (status StatusResponse, err error) {
	var resp statusResponsePayload
	if err = d.request(ctx, "GET", fmt.Sprintf("requests/%s/status", requestID), nil, &resp); err != nil {
		err = fmt.Errorf("failed to perform the HTTP request: %w", err)
	}
	if resp.Code == 0 {
		status = resp.StatusResponse
	} else {
		err = resp.APIError
	}
	return
}

/*
	Result
*/

const (
	RequestStatusSucceeded RequestStatus = "succeeded" // only for Result(), on Status() it is RequestStatusCompleted
	RequestStatusFailed    RequestStatus = "failed"    // only for Result(), on Status() it is RequestStatusCompleted
)

// ResultResponse contains all the information of a completed request.
type ResultResponse struct {
	Status   RequestStatus `json:"status"`
	Created  time.Time     `json:"created_at"`
	Finished time.Time     `json:"finished_at"` // only if status is RequestStatusSucceeded, RequestStatusFailed, RequestStatusCanceled
	Outputs  Outputs       `json:"outputs"`     // only if status is RequestStatusSucceeded
	Error    []ResultError `json:"error"`       // only for RequestStatusFailed
}

// ResultError represents an error that occurred during the execution of a request.
type ResultError struct {
	Code      int    `json:"errorCode"`
	Message   string `json:"error"`
	DebugInfo string `json:"debugInfo"`
}

func (rs *ResultResponse) UnmarshalJSON(data []byte) (err error) {
	// Apply masking to retreive raw values
	type mask ResultResponse
	tmp := struct {
		Created  string `json:"created_at"`
		Finished string `json:"finished_at"`
		*mask
	}{
		mask: (*mask)(rs),
	}
	// Decode
	if err = json.Unmarshal(data, &tmp); err != nil {
		return
	}
	// Convert times
	if rs.Created, err = time.Parse(timeFormat, tmp.Created); err != nil {
		err = fmt.Errorf("failed to parse created time value; %w", err)
		return
	}
	switch tmp.Status {
	case RequestStatusSucceeded, RequestStatusFailed, RequestStatusCanceled:
		if rs.Finished, err = time.Parse(timeFormat, tmp.Finished); err != nil {
			err = fmt.Errorf("failed to parse finished time value %q; %w", tmp.Finished, err)
			return
		}
	}
	return
}

type resultResponsePayload struct {
	ResultResponse
	APIError
}

// Result retrieves the result of a request.
func (d *Deployment) Result(ctx context.Context, requestID string) (result ResultResponse, err error) {
	var resp resultResponsePayload
	if err = d.request(ctx, "GET", fmt.Sprintf("requests/%s/result", requestID), nil, &resp); err != nil {
		err = fmt.Errorf("failed to perform the HTTP request: %w", err)
	}
	if resp.APIError.Code == 0 {
		result = resp.ResultResponse
	} else {
		err = resp.APIError
	}
	return
}

/*
	Cancel
*/

const (
	RequestStatusCancellationRequested RequestStatus = "cancellation_requested" // only for Cancel()
	RequestStatusNotCancellable        RequestStatus = "not_cancellable"        // only for Cancel()
)

type cancelResponsePayload struct {
	RequestID string        `json:"request_id"`
	Status    RequestStatus `json:"status"`
	APIError
}

// Cancel cancels a queued or running job. Returned status can be either RequestStatusCancellationRequested or
// RequestStatusNotCancellable
func (d *Deployment) Cancel(ctx context.Context, requestID string) (status RequestStatus, err error) {
	var resp cancelResponsePayload
	if err = d.request(ctx, "POST", fmt.Sprintf("requests/%s/cancel", requestID), nil, &resp); err != nil {
		err = fmt.Errorf("failed to perform the HTTP request: %w", err)
	}
	if resp.APIError.Code == 0 {
		status = resp.Status
	} else {
		err = resp.APIError
	}
	return
}

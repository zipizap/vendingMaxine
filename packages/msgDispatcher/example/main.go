package main

import (
	msgdispatcher "example/msgDispatcher"
	"fmt"
)

// GetInfoResponse is the response to a GetInfoRequest
type GetInfoResponse struct {
	msgdispatcher.BaseMessage
	Result string `json:"result"`
}

// GetInfoRequest is an example message request type
type GetInfoRequest struct {
	msgdispatcher.BaseMessage
	Query string `json:"query"`
}

// NewGetInfoRequest creates a new GetInfoRequest
func NewGetInfoRequest() *GetInfoRequest {
	return &GetInfoRequest{
		BaseMessage: msgdispatcher.BaseMessage{
			Kind:       "GetInfoRequest",
			ApiVersion: "v1",
		},
	}
}

// Process handles the GetInfoRequest and returns a GetInfoResponse
func (r *GetInfoRequest) Process() msgdispatcher.MessageResponse {
	return &GetInfoResponse{
		BaseMessage: msgdispatcher.BaseMessage{
			Kind:       "GetInfoResponse",
			ApiVersion: r.ApiVersion,
		},
		Result: fmt.Sprintf("Information for query: %s", r.Query),
	}
}

// StatusResponse is the response to a StatusRequest
type StatusResponse struct {
	msgdispatcher.BaseMessage
	Status string `json:"status"`
	Code   int    `json:"code"`
}

// StatusRequest is another example message request type
type StatusRequest struct {
	msgdispatcher.BaseMessage
	System string `json:"system"`
}

// NewStatusRequest creates a new StatusRequest
func NewStatusRequest() *StatusRequest {
	return &StatusRequest{
		BaseMessage: msgdispatcher.BaseMessage{
			Kind:       "StatusRequest",
			ApiVersion: "v1",
		},
	}
}

// Process handles the StatusRequest and returns a StatusResponse
func (r *StatusRequest) Process() msgdispatcher.MessageResponse {
	return &StatusResponse{
		BaseMessage: msgdispatcher.BaseMessage{
			Kind:       "StatusResponse",
			ApiVersion: r.ApiVersion,
		},
		Status: fmt.Sprintf("Status for %s: Online", r.System),
		Code:   200,
	}
}

func main() {
	// Create a new dispatcher
	dispatcher := msgdispatcher.NewDispatcher()

	// Register message types
	dispatcher.Register(NewGetInfoRequest())
	dispatcher.Register(NewStatusRequest())

	// Example 1: Process a GetInfoRequest
	getInfoJSON := []byte(`{"kind": "GetInfoRequest", "apiVersion": "v1", "query": "user profile"}`)
	fmt.Println("Processing:", string(getInfoJSON))

	responseJSON, err := dispatcher.Dispatch(getInfoJSON)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Response:", string(responseJSON))

	// Example 2: Process a StatusRequest
	statusJSON := []byte(`{"kind": "StatusRequest", "apiVersion": "v1", "system": "database"}`)
	fmt.Println("\nProcessing:", string(statusJSON))

	responseJSON, err = dispatcher.Dispatch(statusJSON)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Response:", string(responseJSON))

	// Example 3: Process an unknown request (will produce an error)
	unknownJSON := []byte(`{"kind": "UnknownRequest", "apiVersion": "v1"}`)
	fmt.Println("\nProcessing:", string(unknownJSON))

	responseJSON, err = dispatcher.Dispatch(unknownJSON)
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
	} else {
		fmt.Println("Response:", string(responseJSON))
	}

	// Example 4: Process with incorrect version (will produce an error)
	wrongVersionJSON := []byte(`{"kind": "GetInfoRequest", "apiVersion": "v2", "query": "user profile"}`)
	fmt.Println("\nProcessing:", string(wrongVersionJSON))

	responseJSON, err = dispatcher.Dispatch(wrongVersionJSON)
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
	} else {
		fmt.Println("Response:", string(responseJSON))
	}
}

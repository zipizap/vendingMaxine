package msgdispatcher

import (
	"encoding/json"
	"testing"
)

// GetInfoResponse is a test response
type GetInfoResponse struct {
	BaseMessage
	Result string `json:"result"`
}

// GetInfoRequest is a test request type
type GetInfoRequest struct {
	BaseMessage
	Query string `json:"query"`
}

// NewGetInfoRequest creates a new test request
func NewGetInfoRequest(query string) *GetInfoRequest {
	return &GetInfoRequest{
		BaseMessage: BaseMessage{
			Kind:       "GetInfoRequest",
			ApiVersion: "v1",
		},
		Query: query,
	}
}

// Process handles the test request
func (r *GetInfoRequest) Process() MessageResponse {
	return &GetInfoResponse{
		BaseMessage: BaseMessage{
			Kind:       "GetInfoResponse",
			ApiVersion: r.ApiVersion,
		},
		Result: "Information for query: " + r.Query,
	}
}

// StatusResponse is another test response
type StatusResponse struct {
	BaseMessage
	Status string `json:"status"`
	Code   int    `json:"code"`
}

// StatusRequest is another test request type
type StatusRequest struct {
	BaseMessage
	System string `json:"system"`
}

// NewStatusRequest creates a new test request
func NewStatusRequest(system string) *StatusRequest {
	return &StatusRequest{
		BaseMessage: BaseMessage{
			Kind:       "StatusRequest",
			ApiVersion: "v1",
		},
		System: system,
	}
}

// Process handles the test request
func (r *StatusRequest) Process() MessageResponse {
	return &StatusResponse{
		BaseMessage: BaseMessage{
			Kind:       "StatusResponse",
			ApiVersion: r.ApiVersion,
		},
		Status: "Status for " + r.System + ": Online",
		Code:   200,
	}
}

func TestDispatcher(t *testing.T) {
	// Create a new dispatcher
	dispatcher := NewDispatcher()

	// Register message types
	dispatcher.Register(NewGetInfoRequest(""))
	dispatcher.Register(NewStatusRequest(""))

	// Test GetInfoRequest
	getInfoJSON := `{"kind": "GetInfoRequest", "apiVersion": "v1", "query": "user profile"}`
	responseJSON, err := dispatcher.Dispatch([]byte(getInfoJSON))
	if err != nil {
		t.Fatalf("Failed to dispatch GetInfoRequest: %v", err)
	}

	var getInfoResponse GetInfoResponse
	if err := json.Unmarshal(responseJSON, &getInfoResponse); err != nil {
		t.Fatalf("Failed to unmarshal GetInfoResponse: %v", err)
	}

	if getInfoResponse.Kind != "GetInfoResponse" {
		t.Errorf("Expected kind 'GetInfoResponse', got '%s'", getInfoResponse.Kind)
	}

	if getInfoResponse.ApiVersion != "v1" {
		t.Errorf("Expected ApiVersion 'v1', got '%s'", getInfoResponse.ApiVersion)
	}

	expectedResult := "Information for query: user profile"
	if getInfoResponse.Result != expectedResult {
		t.Errorf("Expected result '%s', got '%s'", expectedResult, getInfoResponse.Result)
	}

	// Test StatusRequest
	statusJSON := `{"kind": "StatusRequest", "apiVersion": "v1", "system": "database"}`
	responseJSON, err = dispatcher.Dispatch([]byte(statusJSON))
	if err != nil {
		t.Fatalf("Failed to dispatch StatusRequest: %v", err)
	}

	var statusResponse StatusResponse
	if err := json.Unmarshal(responseJSON, &statusResponse); err != nil {
		t.Fatalf("Failed to unmarshal StatusResponse: %v", err)
	}

	if statusResponse.Kind != "StatusResponse" {
		t.Errorf("Expected kind 'StatusResponse', got '%s'", statusResponse.Kind)
	}

	if statusResponse.ApiVersion != "v1" {
		t.Errorf("Expected ApiVersion 'v1', got '%s'", statusResponse.ApiVersion)
	}

	expectedStatus := "Status for database: Online"
	if statusResponse.Status != expectedStatus {
		t.Errorf("Expected status '%s', got '%s'", expectedStatus, statusResponse.Status)
	}

	// Test unknown message kind
	unknownJSON := `{"kind": "UnknownRequest", "apiVersion": "v1"}`
	_, err = dispatcher.Dispatch([]byte(unknownJSON))
	if err == nil {
		t.Error("Expected error for unknown message kind, got nil")
	}

	// Test correct kind but wrong apiVersion
	wrongVersionJSON := `{"kind": "GetInfoRequest", "apiVersion": "v2", "query": "test"}`
	_, err = dispatcher.Dispatch([]byte(wrongVersionJSON))
	if err == nil {
		t.Error("Expected error for wrong API version, got nil")
	}
}

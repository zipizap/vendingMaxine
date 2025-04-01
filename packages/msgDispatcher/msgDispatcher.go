package msgdispatcher

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// BaseMessage provides common functionality for all messages
type BaseMessage struct {
	Kind       string `json:"kind"`
	ApiVersion string `json:"apiVersion"`
	Uuid       string `json:"uuid"`
}

// GetKind returns the kind of message
func (b *BaseMessage) GetKind() string {
	return b.Kind
}

// GetApiVersion returns the API version of the message
func (b *BaseMessage) GetApiVersion() string {
	return b.ApiVersion
}

// GetApiVersion returns the API version of the message
func (b *BaseMessage) GetUuid() string {
	return b.Uuid
}

// From a baseMessaRequest, create a corresponding baseMessageResponse
// - kind and apiVersion are the same
// - uuid is the same as the request, but with suffix "Req" replaced with "Resp"
func (bmReq *BaseMessage) CreateBaseMessageResponse() (bmResp *BaseMessage) {
	bmRespKind := strings.TrimSuffix(bmReq.Kind, "Request") + "Response"
	bmRespApiVersion := bmReq.ApiVersion
	bmRespUuid := strings.TrimSuffix(bmReq.Uuid, "Req") + "Resp"
	bmResp = &BaseMessage{
		Kind:       bmRespKind,
		ApiVersion: bmRespApiVersion,
		Uuid:       bmRespUuid,
	}
	return bmResp
}

// BaseRequester interface defines the contract for all request messages
type BaseRequester interface {
	GetKind() string       // implemented by BaseMessage
	GetApiVersion() string // implemented by BaseMessage
	GetUuid() string       // implemented by BaseMessage
	Process() BaseResponser
}

// BaseResponser interface defines the contract for all response messages
type BaseResponser interface {
	GetKind() string       // implemented by BaseMessage
	GetApiVersion() string // implemented by BaseMessage
	GetUuid() string       // implemented by BaseMessage
}

// Dispatcher handles the registration and processing of message types
type Dispatcher struct {
	handlers map[string]reflect.Type // key format: "kind:apiVersion"
}

// NewDispatcher creates a new message dispatcher
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers: make(map[string]reflect.Type),
	}
}

// createHandlerKey creates a composite key from kind and apiVersion
func createHandlerKey(kind, apiVersion string) string {
	return fmt.Sprintf("%s:%s", kind, apiVersion)
}

// Register adds a message request type to the dispatcher
func (d *Dispatcher) Register(messageType BaseRequester) {
	key := createHandlerKey(messageType.GetKind(), messageType.GetApiVersion())
	d.handlers[key] = reflect.TypeOf(messageType).Elem()
}

// Dispatch processes an incoming JSON message and returns a response
// It first determines the message **kind** and **apiVersion**, with which finds the appropriate request type.
// It unmarshals the JSON into the specific request type, processes it, and marshals the response back to JSON
// The requestJSON parameter is the incoming JSON message
// The responseJSON parameter is the outgoing JSON message
func (d *Dispatcher) Dispatch(requestJSON []byte) (responseJSON []byte, err error) {
	// First, determine the message kind and apiVersion
	var baseMsg BaseMessage
	if err := json.Unmarshal(requestJSON, &baseMsg); err != nil {
		return nil, fmt.Errorf("failed to parse message: %w", err)
	}

	// Find the registered handler for this kind and apiVersion
	key := createHandlerKey(baseMsg.Kind, baseMsg.ApiVersion)
	requestType, ok := d.handlers[key]
	if !ok {
		return nil, fmt.Errorf("unknown message kind:version: %s:%s", baseMsg.Kind, baseMsg.ApiVersion)
	}

	// Create a new instance of the appropriate request type
	requestPtr := reflect.New(requestType).Interface().(BaseRequester)

	// Unmarshal the JSON into the specific request type
	if err := json.Unmarshal(requestJSON, requestPtr); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	// Process the request to get the response
	response := requestPtr.Process()

	// Marshal the response to JSON
	responseJSON, err = json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	return responseJSON, nil
}

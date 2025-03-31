# Message Dispatcher

A flexible Go module for handling typed JSON message requests and responses based on their "Kind" and "ApiVersion" fields.

## Features

- Register message types with their processing logic
- Automatic unmarshalling of JSON messages to their correct Go types
- Process messages and return typed responses
- Message routing based on both Kind and ApiVersion
- Simple interface for extending with new message types

## Usage

### 1. Create message types

Define your request and response message types by embedding `BaseMessage`:

```go
// Request type
type MyRequest struct {
    msgdispatcher.BaseMessage
    Data string `json:"data"`
}

// Constructor for convenient registration
func NewMyRequest() *MyRequest {
    return &MyRequest{
        BaseMessage: msgdispatcher.BaseMessage{
            Kind: "MyRequest",
            ApiVersion: "v1",
        },
    }
}

// Process method to handle the request
func (r *MyRequest) Process() msgdispatcher.MessageResponse {
    return &MyResponse{
        BaseMessage: msgdispatcher.BaseMessage{
            Kind: "MyResponse",
            ApiVersion: r.ApiVersion,
        },
        Result: "Processed: " + r.Data,
    }
}

// Response type
type MyResponse struct {
    msgdispatcher.BaseMessage
    Result string `json:"result"`
}
```

### 2. Register and use the dispatcher

```go
// Create a new dispatcher
dispatcher := msgdispatcher.NewDispatcher()

// Register all your message types
dispatcher.Register(NewMyRequest())

// Dispatch an incoming message
jsonMessage := []byte(`{"kind": "MyRequest", "apiVersion": "v1", "data": "test input"}`)
responseJSON, err := dispatcher.Dispatch(jsonMessage)
if err != nil {
    panic(err)
}

// Use the response
fmt.Println(string(responseJSON))
// Output: {"kind":"MyResponse","apiVersion":"v1","result":"Processed: test input"}
```

## Extending with New Message Types

To add a new message type:

1. Define your request struct embedding `BaseMessage`
2. Implement the `Process()` method
3. Define your response struct
4. Register your new request type with the dispatcher

## Version Management

The dispatcher now validates both the "kind" and "apiVersion" fields when routing messages. This allows for:

- Multiple versions of the same message type to coexist
- Graceful API evolution with backward compatibility
- Clear error messages when an unsupported version is received

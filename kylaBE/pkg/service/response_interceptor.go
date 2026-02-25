/**
* Most part of this code is AI generated
**/

package service

import (
	"context"
	"encoding/json"
	"kyla-be/pkg/utils"
	"log"
	"slices"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type ResponseInterceptor struct {
	excludedMethods  []string
	excludedPrefixes []string
	SQSActions       *utils.SQSActions
}

type Message struct {
	Payload          string `json:"payload"`
	Method           string `json:"method"`
	IncomingMetadata string `json:"incoming_metadata"`
	OutgoingMetadata string `json:"outgoing_metadata,omitempty"`
}

// WrappedServerStream needs to be defined

func NewResponseInterceptor(sqsActions *utils.SQSActions) *ResponseInterceptor {
	return &ResponseInterceptor{
		excludedMethods: []string{
			"/grpc.reflection.v1.ServerReflection/ServerReflectionInfo",
			"/grpc.reflection.v1alpha.ServerReflection/ServerReflectionInfo",
		},
		excludedPrefixes: []string{
			"/grpc.reflection.",
			"/grpc.health.v1.Health/",
			"/da.proto.AuthService/ReadAuthContext",
		},
		SQSActions: sqsActions,
	}
}

// WithExcludedMethods allows customizing excluded methods
func (r *ResponseInterceptor) WithExcludedMethods(methods ...string) *ResponseInterceptor {
	r.excludedMethods = append(r.excludedMethods, methods...)
	return r
}

// WithExcludedPrefixes allows customizing excluded prefixes
func (r *ResponseInterceptor) WithExcludedPrefixes(prefixes ...string) *ResponseInterceptor {
	r.excludedPrefixes = append(r.excludedPrefixes, prefixes...)
	return r
}

// shouldLog checks if the method should be logged
func (r *ResponseInterceptor) shouldLog(method string) bool {
	// Check exact method matches
	if slices.Contains(r.excludedMethods, method) {
		return false
	}

	// Check prefix matches
	for _, prefix := range r.excludedPrefixes {
		if strings.HasPrefix(method, prefix) {
			return false
		}
	}

	return true
}

// UnaryServerResponseInterceptor intercepts unary RPC responses
func (r *ResponseInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Extract incoming metadata BEFORE calling handler
		incomingMD, _ := metadata.FromIncomingContext(ctx)

		// Call the handler to get the response
		resp, err := handler(ctx, req)

		// Check if we should log this method
		if r.shouldLog(info.FullMethod) {
			// Extract outgoing metadata (if any was set by the handler)
			outgoingMD, _ := metadata.FromOutgoingContext(ctx)

			// Log response and metadata
			logResponse(r.SQSActions, info.FullMethod, resp, incomingMD, outgoingMD, err)
		}

		return resp, err
	}
}

func (r *ResponseInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := ss.Context()

		// Extract incoming metadata from the stream context
		incomingMD, _ := metadata.FromIncomingContext(ctx)

		// Check if we should log this method
		shouldLog := r.shouldLog(info.FullMethod)

		// Wrap the ServerStream to intercept Send calls
		wrapped := &WrappedServerStream{
			ServerStream:     ss,
			ctx:              ctx,
			method:           info.FullMethod,
			shouldLog:        shouldLog,
			sqsActions:       r.SQSActions,
			incomingMetadata: incomingMD,
		}

		return handler(srv, wrapped)
	}
}

// SendMsg intercepts the response before sending
func (w *WrappedServerStream) SendMsg(m interface{}) error {
	// Only log if this method should be logged
	if w.shouldLog {
		// Extract outgoing metadata (if any)
		outgoingMD, _ := metadata.FromOutgoingContext(w.Context())

		// Log response and metadata
		logResponse(w.sqsActions, w.method, m, w.incomingMetadata, outgoingMD, nil)
	}

	// Send the actual message
	return w.ServerStream.SendMsg(m)
}

// Helper function to log responses with both incoming and outgoing metadata
func logResponse(sqsActions *utils.SQSActions, method string, resp interface{}, incomingMD, outgoingMD metadata.MD, err error) {
	output := map[string]interface{}{
		"method":            method,
		"response":          resp,
		"incoming_metadata": incomingMD,
		"outgoing_metadata": outgoingMD,
	}

	if err != nil {
		output["error"] = err.Error()
	}

	// Send to SQS asynchronously
	if sqsActions != nil {
		protoMsg, ok := resp.(proto.Message)
		if !ok {
			log.Printf("Failed to convert response to proto.Message")
			return
		}

		resJson, err := messageToJSONString(protoMsg)
		if err != nil {
			log.Printf("Failed to marshal proto message: %v", err)
			return
		}

		incomingMetadataJSON, err := json.Marshal(incomingMD)
		if err != nil {
			log.Printf("Failed to marshal incoming metadata: %v", err)
			return
		}

		outgoingMetadataJSON := "{}"
		if len(outgoingMD) > 0 {
			outgoingMetadataBytes, err := json.Marshal(outgoingMD)
			if err != nil {
				log.Printf("Failed to marshal outgoing metadata: %v", err)
			} else {
				outgoingMetadataJSON = string(outgoingMetadataBytes)
			}
		}

		go func() {
			attribute := []utils.MessageAttribute{
				{
					Name:  "method",
					Type:  "String",
					Value: method,
				},
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			queueURL, err := sqsActions.GetQueueUrl(ctx, "service-events")
			if err != nil {
				log.Printf("Failed to get SQS queue URL: %v", err)
				return
			}

			message := &Message{
				Payload:          resJson,
				Method:           method,
				IncomingMetadata: string(incomingMetadataJSON),
				OutgoingMetadata: outgoingMetadataJSON,
			}

			messageBytes, err := json.Marshal(message)
			if err != nil {
				log.Printf("Failed to marshal SQS message: %v", err)
				return
			}

			_, err = sqsActions.SendMessage(ctx, queueURL, string(messageBytes), attribute)
			if err != nil {
				log.Printf("Failed to send message to SQS: %v", err)
			} else {
				log.Printf("Sent message to SQS for method %s", method)
			}
		}()
	}
}

func messageToJSONString(msg proto.Message) (string, error) {
	marshaler := protojson.MarshalOptions{
		Multiline:       false,
		EmitUnpopulated: true,  // include fields with zero values
		UseEnumNumbers:  false, // use enum names
	}
	jsonBytes, err := marshaler.Marshal(msg)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

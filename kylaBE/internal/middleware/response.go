/**
* Most part of this code is AI generated
**/

package middleware

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

// ResponseInterceptor logs gRPC responses to SQS asynchronously.
type ResponseInterceptor struct {
	excludedMethods  []string
	excludedPrefixes []string
	SQSActions       *utils.SQSActions
}

// Message is the payload sent to SQS.
type Message struct {
	Payload          string `json:"payload"`
	Method           string `json:"method"`
	IncomingMetadata string `json:"incoming_metadata"`
	OutgoingMetadata string `json:"outgoing_metadata,omitempty"`
}

// NewResponseInterceptor creates a ResponseInterceptor.
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

// WithExcludedMethods appends extra methods to the exclusion list.
func (r *ResponseInterceptor) WithExcludedMethods(methods ...string) *ResponseInterceptor {
	r.excludedMethods = append(r.excludedMethods, methods...)
	return r
}

// WithExcludedPrefixes appends extra prefixes to the exclusion list.
func (r *ResponseInterceptor) WithExcludedPrefixes(prefixes ...string) *ResponseInterceptor {
	r.excludedPrefixes = append(r.excludedPrefixes, prefixes...)
	return r
}

func (r *ResponseInterceptor) shouldLog(method string) bool {
	if slices.Contains(r.excludedMethods, method) {
		return false
	}
	for _, prefix := range r.excludedPrefixes {
		if strings.HasPrefix(method, prefix) {
			return false
		}
	}
	return true
}

// Unary returns the unary response interceptor.
func (r *ResponseInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		incomingMD, _ := metadata.FromIncomingContext(ctx)
		resp, err := handler(ctx, req)
		if r.shouldLog(info.FullMethod) {
			outgoingMD, _ := metadata.FromOutgoingContext(ctx)
			logResponse(r.SQSActions, info.FullMethod, resp, incomingMD, outgoingMD, err)
		}
		return resp, err
	}
}

// Stream returns the streaming response interceptor.
func (r *ResponseInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := ss.Context()
		incomingMD, _ := metadata.FromIncomingContext(ctx)
		shouldLog := r.shouldLog(info.FullMethod)
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

// SendMsg intercepts the outgoing message before forwarding it.
func (w *WrappedServerStream) SendMsg(m interface{}) error {
	if w.shouldLog {
		outgoingMD, _ := metadata.FromOutgoingContext(w.Context())
		logResponse(w.sqsActions, w.method, m, w.incomingMetadata, outgoingMD, nil)
	}
	return w.ServerStream.SendMsg(m)
}

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
			if b, err := json.Marshal(outgoingMD); err == nil {
				outgoingMetadataJSON = string(b)
			}
		}
		go func() {
			attribute := []utils.MessageAttribute{
				{Name: "method", Type: "String", Value: method},
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
		EmitUnpopulated: true,
		UseEnumNumbers:  false,
	}
	jsonBytes, err := marshaler.Marshal(msg)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

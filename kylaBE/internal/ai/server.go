package ai

import (
	"context"
	"log"

	"kyla-be/pkg/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements pb.AIServiceServer. The actual model work is delegated to
// an LLMProvider so the gRPC layer stays a thin adapter — the same provider
// instance is also handed to the automation worker so workflows and gRPC
// callers share configuration.
type Server struct {
	provider LLMProvider
	pb.UnimplementedAIServiceServer
}

func NewServer(provider LLMProvider) *Server {
	if provider == nil {
		provider = NoopProvider{}
	}
	return &Server{provider: provider}
}

func (s *Server) ClassifyText(ctx context.Context, req *pb.ClassifyTextRequest) (*pb.ClassifyTextResponse, error) {
	if req.GetText() == "" || len(req.GetLabels()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "text and labels are required")
	}
	label, conf, err := s.provider.Classify(ctx, req.GetText(), req.GetLabels())
	if err != nil {
		log.Printf("[ai] classify (%s): %v", s.provider.Name(), err)
		return nil, status.Error(codes.Internal, "classify failed")
	}
	return &pb.ClassifyTextResponse{Label: label, Confidence: conf}, nil
}

func (s *Server) SummarizeText(ctx context.Context, req *pb.SummarizeTextRequest) (*pb.SummarizeTextResponse, error) {
	if req.GetText() == "" {
		return nil, status.Error(codes.InvalidArgument, "text is required")
	}
	summary, err := s.provider.Summarize(ctx, req.GetText(), int(req.GetMaxSentences()))
	if err != nil {
		log.Printf("[ai] summarize (%s): %v", s.provider.Name(), err)
		return nil, status.Error(codes.Internal, "summarize failed")
	}
	return &pb.SummarizeTextResponse{Summary: summary}, nil
}

func (s *Server) GenerateReply(ctx context.Context, req *pb.GenerateReplyRequest) (*pb.GenerateReplyResponse, error) {
	if req.GetPrompt() == "" {
		return nil, status.Error(codes.InvalidArgument, "prompt is required")
	}
	reply, err := s.provider.GenerateReply(ctx, req.GetHistory(), req.GetPrompt())
	if err != nil {
		log.Printf("[ai] generate_reply (%s): %v", s.provider.Name(), err)
		return nil, status.Error(codes.Internal, "generate_reply failed")
	}
	return &pb.GenerateReplyResponse{Reply: reply}, nil
}

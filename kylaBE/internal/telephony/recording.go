package telephony

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Transcriber is the audio→text capability the recording pipeline needs.
// internal/ai supplies a concrete implementation; defining the interface
// here keeps telephony from importing internal/ai (mirrors the
// `activities.AIClassifier` pattern from Phase 6).
type Transcriber interface {
	Name() string
	TranscribeAudio(ctx context.Context, audio []byte, mime string) (string, error)
}

// RecordingPipeline runs the post-call recording workflow:
//
//	RECORD_STOP event arrives  →
//	  persist CallRecording row (status=pending,pending)         →
//	  read the PBX-local audio file (mounted via freeswitch_recordings volume) →
//	  upload to S3 (status=uploaded)                            →
//	  transcribe via Transcriber (status=done)                  →
//	  rebuild call.transcript from all recordings of the call.
//
// One pipeline per binary. Each recording is processed in its own goroutine
// so a slow upload doesn't block other recordings.
type RecordingPipeline struct {
	store        *Store
	uploader     *manager.Uploader
	bucket       string
	keyPrefix    string                  // optional path prefix in the bucket
	transcriber  Transcriber
	httpEndpoint string                  // optional S3 endpoint override (MinIO, etc.)
}

// RecordingPipelineConfig groups the runtime knobs.
type RecordingPipelineConfig struct {
	AWSRegion    string
	AWSAccessKey string
	AWSSecretKey string
	Bucket       string
	KeyPrefix    string
	S3Endpoint   string // optional, e.g. "https://minio.local:9000" for self-hosted S3
}

// NewRecordingPipeline constructs the pipeline. Returns (nil, nil) when the
// bucket isn't configured — call sites then skip enqueuing recordings.
// Errors are returned only on misconfiguration (invalid credentials etc.).
func NewRecordingPipeline(ctx context.Context, cfg RecordingPipelineConfig, store *Store, transcriber Transcriber) (*RecordingPipeline, error) {
	if cfg.Bucket == "" {
		log.Println("recording pipeline: bucket not configured; pipeline disabled")
		return nil, nil
	}
	if store == nil {
		return nil, errors.New("recording pipeline: store required")
	}

	awsConfig, err := awscfg.LoadDefaultConfig(ctx,
		awscfg.WithRegion(cfg.AWSRegion),
		awscfg.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AWSAccessKey, cfg.AWSSecretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("recording pipeline: aws config: %w", err)
	}

	s3Opts := []func(*s3.Options){}
	if cfg.S3Endpoint != "" {
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.S3Endpoint)
			o.UsePathStyle = true
		})
	}
	s3Client := s3.NewFromConfig(awsConfig, s3Opts...)
	uploader := manager.NewUploader(s3Client)

	return &RecordingPipeline{
		store:        store,
		uploader:     uploader,
		bucket:       cfg.Bucket,
		keyPrefix:    strings.Trim(cfg.KeyPrefix, "/"),
		transcriber:  transcriber,
		httpEndpoint: cfg.S3Endpoint,
	}, nil
}

// Enabled reports whether the pipeline can do real work. The EventBridge
// short-circuits when Enabled() is false.
func (p *RecordingPipeline) Enabled() bool {
	return p != nil && p.uploader != nil && p.bucket != ""
}

// Handle records the RECORD_STOP event into a CallRecording row and kicks
// off the upload+transcribe pipeline in the background. Returns immediately —
// the bridge stays responsive.
//
// callID is the PBX UUID for the call leg the recording belongs to; pbxPath
// is the on-disk path on the FreeSWITCH container (also accessible to the
// Go process via the shared freeswitch_recordings volume).
func (p *RecordingPipeline) Handle(ctx context.Context, callID, orgID, pbxPath string) {
	if !p.Enabled() {
		return
	}
	if callID == "" || pbxPath == "" {
		return
	}
	rec := &CallRecording{
		CallID:       callID,
		OrgID:        orgID,
		PBXPath:      pbxPath,
		UploadStatus: string(UploadPending),
		TranscribeStatus: string(TranscribePending),
		RecordedAt:   time.Now().UTC(),
	}
	created, err := p.store.CreateRecording(rec)
	if err != nil {
		log.Printf("[recording] persist row failed (call=%s): %v", callID, err)
		return
	}
	_ = p.store.SetTranscriptStatus(callID, "pending", "")
	go p.process(context.Background(), created)
}

// process runs the upload + transcribe pipeline for a single recording.
// Errors are logged and recorded on the row but never crash the goroutine.
func (p *RecordingPipeline) process(ctx context.Context, rec *CallRecording) {
	// 1. Upload to S3.
	if err := p.upload(ctx, rec); err != nil {
		log.Printf("[recording] upload failed (rec=%s): %v", rec.ID, err)
		_ = p.store.SetRecordingUploadFailed(rec.ID, err.Error())
		_ = p.store.SetTranscriptStatus(rec.CallID, string(TranscribeFailed), "upload_failed")
		return
	}

	// 2. Transcribe.
	if p.transcriber == nil {
		_ = p.store.SetRecordingTranscribed(rec.ID, "", "noop")
		_ = p.store.SetTranscriptStatus(rec.CallID, "done", "")
		return
	}
	_ = p.store.SetTranscriptStatus(rec.CallID, string(TranscribeRunning), "")
	if err := p.transcribe(ctx, rec); err != nil {
		log.Printf("[recording] transcribe failed (rec=%s): %v", rec.ID, err)
		_ = p.store.SetRecordingTranscribeFailed(rec.ID, err.Error())
		_ = p.store.SetTranscriptStatus(rec.CallID, string(TranscribeFailed), err.Error())
		return
	}

	// 3. Rebuild the call-level transcript from all recordings.
	if err := p.refreshCallTranscript(rec.CallID); err != nil {
		log.Printf("[recording] refresh call transcript failed (call=%s): %v", rec.CallID, err)
	}
}

// upload reads the PBX-local file and ships it to S3 via the SDK uploader.
// On success the recording row is updated with bucket+key+url+size.
func (p *RecordingPipeline) upload(ctx context.Context, rec *CallRecording) error {
	f, err := os.Open(rec.PBXPath)
	if err != nil {
		return fmt.Errorf("open %s: %w", rec.PBXPath, err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return fmt.Errorf("stat: %w", err)
	}

	// Compose key: <prefix>/<org_id>/<call_id>/<recording_id>.<ext>.
	ext := strings.TrimPrefix(filepath.Ext(rec.PBXPath), ".")
	if ext == "" {
		ext = "wav"
	}
	parts := []string{}
	if p.keyPrefix != "" {
		parts = append(parts, p.keyPrefix)
	}
	parts = append(parts, rec.OrgID, rec.CallID, rec.ID+"."+ext)
	key := strings.Join(parts, "/")

	if _, err := p.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(p.bucket),
		Key:         aws.String(key),
		Body:        f,
		ContentType: aws.String(rec.ContentType),
	}); err != nil {
		return fmt.Errorf("s3 put: %w", err)
	}

	url := p.objectURL(key)
	if err := p.store.SetRecordingUploaded(rec.ID, p.bucket, key, url, stat.Size()); err != nil {
		return fmt.Errorf("persist uploaded state: %w", err)
	}
	_ = p.store.SetRecordingURL(rec.CallID, url) // surface the most recent recording on the call row
	return nil
}

// transcribe reads the local file (same path the upload used) and ships its
// bytes to the Transcriber. Using the local file rather than re-downloading
// from S3 keeps the path short and avoids egress.
func (p *RecordingPipeline) transcribe(ctx context.Context, rec *CallRecording) error {
	bytes, err := os.ReadFile(rec.PBXPath)
	if err != nil {
		return fmt.Errorf("read for transcription: %w", err)
	}
	transcript, err := p.transcriber.TranscribeAudio(ctx, bytes, rec.ContentType)
	if err != nil {
		return fmt.Errorf("transcribe: %w", err)
	}
	return p.store.SetRecordingTranscribed(rec.ID, transcript, p.transcriber.Name())
}

// refreshCallTranscript concatenates all per-recording transcripts in
// recorded_at order and stores the result on the call row. Idempotent.
func (p *RecordingPipeline) refreshCallTranscript(callID string) error {
	recs, err := p.store.ListRecordingsForCall(callID)
	if err != nil {
		return err
	}
	var parts []string
	provider := ""
	for _, r := range recs {
		if r.TranscribeStatus != string(TranscribeDone) {
			continue
		}
		parts = append(parts, r.Transcript)
		if r.TranscribedBy != "" {
			provider = r.TranscribedBy
		}
	}
	return p.store.SetTranscript(callID, strings.Join(parts, "\n\n"), provider)
}

// objectURL returns a stable URL for the uploaded object. When a custom
// endpoint is configured we use path-style; otherwise we use the public
// virtual-hosted-style URL.
func (p *RecordingPipeline) objectURL(key string) string {
	if p.httpEndpoint != "" {
		return strings.TrimRight(p.httpEndpoint, "/") + "/" + p.bucket + "/" + key
	}
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", p.bucket, key)
}

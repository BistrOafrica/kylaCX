package telephony

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RecordingUploadStatus enumerates the upload pipeline's terminal states.
type RecordingUploadStatus string

const (
	UploadPending   RecordingUploadStatus = "pending"
	UploadUploading RecordingUploadStatus = "uploading"
	UploadUploaded  RecordingUploadStatus = "uploaded"
	UploadFailed    RecordingUploadStatus = "failed"
)

// RecordingTranscribeStatus enumerates the transcription pipeline's states.
type RecordingTranscribeStatus string

const (
	TranscribePending RecordingTranscribeStatus = "pending"
	TranscribeRunning RecordingTranscribeStatus = "running"
	TranscribeDone    RecordingTranscribeStatus = "done"
	TranscribeFailed  RecordingTranscribeStatus = "failed"
	TranscribeSkipped RecordingTranscribeStatus = "skipped"
)

// CallRecording is one captured audio file. A call may have multiple
// recordings (one per leg, one per IVR record node, etc.); aggregation onto
// the calls row happens after each recording finishes transcription.
type CallRecording struct {
	ID     string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CallID string `gorm:"type:uuid;not null;index" json:"call_id"`
	OrgID  string `gorm:"type:uuid;not null;index" json:"org_id"`

	PBXPath  string `gorm:"not null" json:"pbx_path"`
	S3Bucket string `gorm:"not null;default:''" json:"s3_bucket,omitempty"`
	S3Key    string `gorm:"not null;default:''" json:"s3_key,omitempty"`
	S3URL    string `gorm:"not null;default:''" json:"s3_url,omitempty"`

	DurationSeconds int    `gorm:"not null;default:0" json:"duration_seconds"`
	SizeBytes       int64  `gorm:"not null;default:0" json:"size_bytes"`
	ContentType     string `gorm:"not null;default:'audio/wav'" json:"content_type"`

	UploadStatus string `gorm:"not null;default:'pending';index" json:"upload_status"`
	UploadError  string `gorm:"not null;default:''" json:"upload_error,omitempty"`

	TranscribeStatus string `gorm:"not null;default:'pending';index" json:"transcribe_status"`
	TranscribeError  string `gorm:"not null;default:''" json:"transcribe_error,omitempty"`
	Transcript       string `gorm:"not null;default:''" json:"transcript,omitempty"`
	TranscribedBy    string `gorm:"not null;default:''" json:"transcribed_by,omitempty"`

	RecordedAt     time.Time  `gorm:"not null;default:now()" json:"recorded_at"`
	UploadedAt     *time.Time `json:"uploaded_at,omitempty"`
	TranscribedAt  *time.Time `json:"transcribed_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (CallRecording) TableName() string { return "call_recordings" }

func (r *CallRecording) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.NewString()
	}
	return nil
}

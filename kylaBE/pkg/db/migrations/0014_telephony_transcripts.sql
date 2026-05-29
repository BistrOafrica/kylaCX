-- Phase 5e — Call recording uploads + transcripts
-- Adds:
--   * transcript columns on calls (transcribed text + transcription status)
--   * recordings table for the upload pipeline (one row per recording file)

ALTER TABLE calls
  ADD COLUMN IF NOT EXISTS transcript           TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS transcript_status    TEXT NOT NULL DEFAULT 'none',
  -- pending | uploading | uploaded | transcribing | done | failed | none
  ADD COLUMN IF NOT EXISTS transcript_provider  TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS transcript_error     TEXT NOT NULL DEFAULT '';

-- One row per recording file produced by the PBX. We persist the PBX-local
-- path immediately on RECORD_STOP, then the uploader moves the file to S3
-- and the transcriber consumes the S3 URI. Keeping a dedicated table lets
-- the same call have multiple recordings (e.g. one per leg, or per IVR
-- record node) without overloading the calls row.
CREATE TABLE IF NOT EXISTS call_recordings (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  call_id       UUID NOT NULL,
  org_id        UUID NOT NULL,

  pbx_path      TEXT NOT NULL,                                -- e.g. /var/lib/freeswitch/recordings/<uuid>.wav
  s3_bucket     TEXT NOT NULL DEFAULT '',
  s3_key        TEXT NOT NULL DEFAULT '',
  s3_url        TEXT NOT NULL DEFAULT '',

  duration_seconds INTEGER NOT NULL DEFAULT 0,
  size_bytes       BIGINT  NOT NULL DEFAULT 0,
  content_type     TEXT    NOT NULL DEFAULT 'audio/wav',

  -- Upload state: pending | uploading | uploaded | failed
  upload_status TEXT NOT NULL DEFAULT 'pending',
  upload_error  TEXT NOT NULL DEFAULT '',

  -- Transcription state: pending | running | done | failed | skipped
  transcribe_status TEXT NOT NULL DEFAULT 'pending',
  transcribe_error  TEXT NOT NULL DEFAULT '',
  transcript        TEXT NOT NULL DEFAULT '',
  transcribed_by    TEXT NOT NULL DEFAULT '',                 -- provider name (e.g. "openai-whisper")

  recorded_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  uploaded_at  TIMESTAMPTZ,
  transcribed_at TIMESTAMPTZ,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_call_recordings_call   ON call_recordings (call_id);
CREATE INDEX IF NOT EXISTS idx_call_recordings_states ON call_recordings (upload_status, transcribe_status);

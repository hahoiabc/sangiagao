-- Migration 016: Call logs table for voice/video call history
CREATE TABLE IF NOT EXISTS call_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id),
    caller_id UUID NOT NULL REFERENCES users(id),
    callee_id UUID NOT NULL REFERENCES users(id),
    call_type VARCHAR(10) NOT NULL DEFAULT 'audio' CHECK (call_type IN ('audio', 'video')),
    status VARCHAR(20) NOT NULL DEFAULT 'missed' CHECK (status IN ('missed', 'answered', 'rejected', 'busy', 'failed')),
    duration_seconds INT NOT NULL DEFAULT 0,
    started_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_call_logs_conversation ON call_logs(conversation_id, created_at DESC);
CREATE INDEX idx_call_logs_caller ON call_logs(caller_id, created_at DESC);
CREATE INDEX idx_call_logs_callee ON call_logs(callee_id, created_at DESC);

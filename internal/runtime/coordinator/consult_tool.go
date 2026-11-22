package coordinator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"reasonix/internal/contract/tool"
)

const (
	ConsultationToolName    = "consult_planner"
	consultationQuestionMax = 2000
	consultationEvidenceMax = 8000
)

var (
	ErrConsultationUnavailable         = errors.New("planner consultation unavailable")
	ErrConsultationLimit               = errors.New("planner consultation limit reached")
	ErrConsultationTimeout             = errors.New("planner consultation timed out")
	ErrConsultationCancelled           = errors.New("planner consultation cancelled")
	ErrConsultationApprovalUnavailable = errors.New("planner consultation approval unavailable")
)

func consultationFailure(identity error, code, message string) error {
	return fmt.Errorf("%w: %w", identity, tool.Refusal{Code: code, Message: message})
}

type consultationRequest struct {
	Question string `json:"question"`
	Evidence string `json:"evidence,omitempty"`
}

type consultationTool struct {
	coordinator *Coordinator
	consult     func(context.Context, string, string) (string, error)
}

// NewConsultationTool exposes the fixed consultation schema to an executor
// registry for a session that opted in at its boundary.
func NewConsultationTool(c *Coordinator) tool.Tool { return consultationTool{coordinator: c} }

func (consultationTool) Name() string { return ConsultationToolName }
func (consultationTool) Description() string {
	return "Ask the active planner for targeted advice or a validated plan revision. The host enforces consultation limits and approval."
}
func (consultationTool) Schema() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{"question":{"type":"string","minLength":1,"maxLength":2000},"evidence":{"type":"string","maxLength":8000}},"required":["question"],"additionalProperties":false}`)
}
func (consultationTool) ReadOnly() bool { return false }
func (t consultationTool) Execute(ctx context.Context, raw json.RawMessage) (string, error) {
	var req consultationRequest
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}
	req.Question = strings.TrimSpace(req.Question)
	req.Evidence = strings.TrimSpace(req.Evidence)
	if req.Question == "" || utf8.RuneCountInString(req.Question) > consultationQuestionMax {
		return "", fmt.Errorf("question must contain 1 to %d characters", consultationQuestionMax)
	}
	if utf8.RuneCountInString(req.Evidence) > consultationEvidenceMax {
		return "", fmt.Errorf("evidence must be at most %d characters", consultationEvidenceMax)
	}
	if t.consult != nil {
		return t.consult(ctx, req.Question, req.Evidence)
	}
	if t.coordinator == nil {
		return "", ErrConsultationUnavailable
	}
	return t.coordinator.Consult(ctx, req.Question, req.Evidence)
}

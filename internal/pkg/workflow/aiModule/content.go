package aiModule

import (
	"ggb_server/internal/consts"
)

type Type string

const (
	Classify      Type = "classify"
	Reasoning     Type = "reasoning"
	OutputContent Type = "outputContent"
	Element       Type = "element"
	Error         Type = "error"
	Complete      Type = "complete"
	HTMLCode      Type = "html_code"
)

type Content struct {
	Type                   Type               `json:"type"`
	Content                string             `json:"content"`
	Step                   consts.ProcessStep `json:"step"`
	TimeStamp              int64              `json:"time_stamp"`
	ThinkingCostTime       int64              `json:"thinking_cost_time"`
	RootFlowId             int                `json:"root_flow_id"`
	ParentFlowId           int                `json:"parent_flow_id"`
	RelatedSessionId       int                `json:"related_session_id"`
	RelatedMessageId       int                `json:"related_message_id"`
	ParentRelatedMessageId int                `json:"parent_related_message_id"`
}

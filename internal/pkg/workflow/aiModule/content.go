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
	Type                   Type
	Content                string
	Step                   consts.ProcessStep
	TimeStamp              int64
	ThinkingCostTime       int64
	RootFlowId             int
	ParentFlowId           int
	RelatedSessionId       int
	RelatedMessageId       int
	ParentRelatedMessageId int
}

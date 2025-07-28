package models

type N8NMainWorkflowResponse []N8NMainWorkflowResponseItem

type N8NMainWorkflowResponseItem struct {
	Answer    string `json:"answer"`
	ShouldEnd string `json:"should_end"`
}

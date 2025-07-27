package models

type N8NMainWorkflowResponse []N8NMainWorkflowResponseItem

type N8NMainWorkflowResponseItem struct {
	Answer     string `json:"answer"`
	DialNumber string `json:"dial_number"`
	ShouldEnd  string `json:"should_end"`
}

package models

type OCRRequest struct {
	Version   string     `json:"version"`
	RequestId string     `json:"requestId"`
	Timestamp string     `json:"timestamp"`
	Lang      string     `json:"lang, omitempty"`
	Images    []OCRImage `json:"images"`
}

type OCRImage struct {
	Format     string   `json:"format"`
	Data       string   `json:"data,omitempty"`
	Url        string   `json:"url,omitempty"`
	Name       string   `json:"name"`
	TemplateId []string `json:"templateId, omitempty"`
}

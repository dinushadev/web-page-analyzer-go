package model

type HeadingCount struct {
	Level int `json:"level"`
	Count int `json:"count"`
}

type LinkStats struct {
	Internal int `json:"internal"`
	External int `json:"external"`
	Inaccessible int `json:"inaccessible"`
}

type AnalyzeResult struct {
	HTMLVersion string        `json:"html_version"`
	Title       string        `json:"title"`
	Headings    []HeadingCount `json:"headings"`
	Links       LinkStats     `json:"links"`
	LoginForm   bool          `json:"login_form"`
}

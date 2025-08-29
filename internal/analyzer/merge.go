package analyzer

import "test-project-go/internal/model"

func mergeAnalyzeResult(main, partial *model.AnalyzeResult) {
	if partial.HTMLVersion != "" {
		main.HTMLVersion = partial.HTMLVersion
	}
	if partial.Title != "" {
		main.Title = partial.Title
	}
	if len(partial.Headings) > 0 {
		main.Headings = partial.Headings
	}
	if partial.Links.Internal != 0 || partial.Links.External != 0 || partial.Links.Inaccessible != 0 {
		main.Links = partial.Links
	}
	if partial.LoginForm {
		main.LoginForm = true
	}
}

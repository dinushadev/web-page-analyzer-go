package service

import (
	"net/url"
	"golang.org/x/net/html"
	"test-project-go/internal/model"
)

type AnalyzerStrategy interface {
	Analyze(doc *html.Node, base *url.URL, result *model.AnalyzeResult) error
}

type HTMLVersionStrategy struct{}
func (s *HTMLVersionStrategy) Analyze(doc *html.Node, base *url.URL, result *model.AnalyzeResult) error {
	result.HTMLVersion = detectHTMLVersion(doc)
	return nil
}

type TitleStrategy struct{}
func (s *TitleStrategy) Analyze(doc *html.Node, base *url.URL, result *model.AnalyzeResult) error {
	result.Title = extractTitle(doc)
	return nil
}

type HeadingsStrategy struct{}
func (s *HeadingsStrategy) Analyze(doc *html.Node, base *url.URL, result *model.AnalyzeResult) error {
	result.Headings = countHeadings(doc)
	return nil
}

type LinksStrategy struct{
	LinkChecker LinkChecker
}
func (s *LinksStrategy) Analyze(doc *html.Node, base *url.URL, result *model.AnalyzeResult) error {
	internal, external, inaccessible := countLinks(doc, base, s.LinkChecker)
	result.Links = model.LinkStats{Internal: internal, External: external, Inaccessible: inaccessible}
	return nil
}

type LoginFormStrategy struct{}
func (s *LoginFormStrategy) Analyze(doc *html.Node, base *url.URL, result *model.AnalyzeResult) error {
	result.LoginForm = hasLoginForm(doc)
	return nil
}

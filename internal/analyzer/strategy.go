package analyzer

import (
	"net/url"
	"web-analyzer-go/internal/factory"
	"web-analyzer-go/internal/model"
	"web-analyzer-go/internal/util"

	"golang.org/x/net/html"
)

type AnalyzerStrategy interface {
	Analyze(doc *html.Node, base *url.URL, result *model.AnalyzeResult) error
}

type HTMLVersionStrategy struct{}

func (s *HTMLVersionStrategy) Analyze(doc *html.Node, base *url.URL, result *model.AnalyzeResult) error {
	result.HTMLVersion = util.DetectHTMLVersion(doc)
	return nil
}

type TitleStrategy struct{}

func (s *TitleStrategy) Analyze(doc *html.Node, base *url.URL, result *model.AnalyzeResult) error {
	result.Title = util.ExtractTitle(doc)
	return nil
}

type HeadingsStrategy struct{}

func (s *HeadingsStrategy) Analyze(doc *html.Node, base *url.URL, result *model.AnalyzeResult) error {
	result.Headings = util.CountHeadings(doc)
	return nil
}

type LinksStrategy struct {
	LinkChecker factory.LinkChecker
}

func (s *LinksStrategy) Analyze(doc *html.Node, base *url.URL, result *model.AnalyzeResult) error {
	internal, external, inaccessible := util.CountLinks(doc, base, s.LinkChecker.IsAccessible)
	result.Links = model.LinkStats{Internal: internal, External: external, Inaccessible: inaccessible}
	return nil
}

type LoginFormStrategy struct{}

func (s *LoginFormStrategy) Analyze(doc *html.Node, base *url.URL, result *model.AnalyzeResult) error {
	result.LoginForm = util.HasLoginForm(doc)
	return nil
}

package analyzer

import "errors"

var ErrUnreachable = errors.New("url is unreachable")
var ErrInvalidURL = errors.New("invalid url")
var ErrUpstream = errors.New("upstream http error")
var ErrParseHTML = errors.New("failed to parse html")

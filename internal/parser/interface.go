package parser

type Parser interface {
	Parse(path string) (TaskMap, error)
}

type DefaultParser struct{}

func NewParser() Parser {
	return &DefaultParser{}
}

func (p *DefaultParser) Parse(path string) (TaskMap, error) {
	return Parse(path)
}

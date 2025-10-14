package finder

type Finder interface {
	FindBabfile() (string, error)
}

type DefaultFinder struct{}

func NewFinder() Finder {
	return &DefaultFinder{}
}

func (f *DefaultFinder) FindBabfile() (string, error) {
	return FindBabfile()
}

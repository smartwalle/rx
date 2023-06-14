package rx

import "net/http"

type Provider interface {
	Match(req *http.Request) (*Location, error)
}

type ListProvider struct {
	locations []*Location
}

func NewListProvider() *ListProvider {
	return &ListProvider{}
}

func (this *ListProvider) Match(req *http.Request) (*Location, error) {
	var path = req.URL.Path
	for _, location := range this.locations {
		if location.Match(path) {
			return location, nil
		}
	}
	return nil, nil
}

func (this *ListProvider) Add(pattern string, targets []string, opts ...Option) error {
	var location, err = NewLocation(pattern, targets, opts...)
	if err != nil {
		return err
	}
	this.locations = append(this.locations, location)
	return nil
}

package rx

type Params map[string]string

func (p Params) Get(key string) string {
	return p[key]
}

func (p Params) Set(key, value string) {
	p[key] = value
}

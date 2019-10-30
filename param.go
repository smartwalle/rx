package rx

type Param struct {
	key   string
	value string
}

type Params []Param

func (ps Params) Get(key string) string {
	for _, p := range ps {
		if p.key == key {
			return p.value
		}
	}
	return ""
}

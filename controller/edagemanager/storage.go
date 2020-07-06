package edagemanager

type rangeFunc func(key string, val *Edage) bool

type IStorage interface {
	Set(key string, val *Edage)
	Get(key string) *Edage
	Del(key string)
	List() map[string]*Edage
	Range(funcCall rangeFunc)
}

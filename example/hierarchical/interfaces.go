package hierarchical

type Service interface {
	Start()
	Stop()
}

type Store interface {
	StoreTransanction(tx string)
}

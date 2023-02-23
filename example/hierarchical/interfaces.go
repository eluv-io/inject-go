package hierarchical

type Service interface {
	Start()
	Stop()
}

type Store interface {
	StoreTransaction(tx string)
}

package storage

type Storage interface {
	Users() UserRepository
	Orders() OrderRepository
}

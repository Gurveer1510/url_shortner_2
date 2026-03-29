package session

type Store interface {
	Create(userId, email string) (*Session, error)
	Get(id string) (*Session, bool)
	Delete(id string)
}

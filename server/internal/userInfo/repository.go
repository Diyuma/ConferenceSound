package userInfo

type Repository interface {
	SetId(string, uint64) error // key value timeExperation
	GetId(string) (bool, uint64, error)
}

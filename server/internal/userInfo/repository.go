package userInfo

type Repository interface {
	SetBitRate(string, int) error // key value timeExperation
	GetBitRate(string) (bool, int, error)
}

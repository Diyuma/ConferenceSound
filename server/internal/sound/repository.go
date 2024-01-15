package sound

import "time"

type Repository interface {
	SetSound(string, *Sound, time.Duration) error // key value timeExperation
	GetSound(string) (*Sound, error)
	GetDelSound(string) (*Sound, error)
}

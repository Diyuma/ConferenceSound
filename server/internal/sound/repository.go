package sound

import "time"

// TODO move to soundadapters folder

type Repository interface {
	SetSound(string, *Sound, time.Duration) error // key value timeExperation
	GetSound(string) (bool, *Sound, error)
	GetDelSound(string) (bool, *Sound, error)
}

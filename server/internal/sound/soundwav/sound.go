package soundwav

import (
	"errors"
	"homework/server/internal/sound"
)

type SoundWav struct {
	data     *[]float32
	bitRate  int
	duration int
	authors  []uint32
	timeSend []uint64
}

var ErrorNotEqualSoundDuration error = errors.New("incorrect sound wave duration - they must be equal")
var ErrorIncorrectSoundDuration error = errors.New("can't divide sound into parts as it has wrong duration or bitRate")
var ErrorIncorrectSoundType error = errors.New("Incorrect sound type in divide into parts")

func NewSound(data *[]float32, bitRate int, duration int, authors []uint32, timeSend []uint64) *SoundWav {
	return &SoundWav{data: data, bitRate: bitRate, duration: duration, authors: authors, timeSend: timeSend}
}

func NewEmptySound() sound.Sound {
	return &SoundWav{}
}

func (s *SoundWav) NewEmptySound() sound.Sound {
	return &SoundWav{data: nil, bitRate: 0, duration: 0, authors: []uint32{}, timeSend: []uint64{}}
}

func (s *SoundWav) NewSound(data *[]float32, bitRate int, duration int, authors []uint32, timeSend []uint64) sound.Sound {
	return &SoundWav{data: nil, bitRate: 0, duration: 0, authors: []uint32{}, timeSend: []uint64{}}
}

func (s *SoundWav) GetData() *[]float32 {
	return s.data
}

func (s *SoundWav) GetAuthors() []uint32 {
	return s.authors
}

func (s *SoundWav) GetSoundDuration() int {
	return len(*s.data) / s.bitRate
}

func (s *SoundWav) GetBitRate() int {
	return s.bitRate
}

func (s *SoundWav) GetTimeId() []uint64 {
	return s.timeSend
}

func (s *SoundWav) SetTimeId(id uint64) {
	s.timeSend = append(s.timeSend, id)
}

// TODO check what is better - to use avarage or just take every kth
func (s *SoundWav) RebitSound(newBitRate int) {
	if s.bitRate > newBitRate {
		k := s.bitRate / newBitRate
		for i := 0; i*k < len(*s.data); i++ {
			(*s.data)[i] = (*s.data)[i*k]
		}

		(*s.data) = (*s.data)[:len(*s.data)/k]
	} else {
		// it is very bad for sound quality if I understand correctly
		k := newBitRate / s.bitRate
		(*s.data) = append(*s.data, make([]float32, len(*s.data)*(k-1))...)
		for i := len(*s.data)/k - 1; i >= 0; i-- {
			for j := 0; j < k; j++ {
				(*s.data)[i*k+j] = (*s.data)[i]
			}
		}
	}
}

func (s *SoundWav) Add(s2 *sound.Sound) error {
	if len(*s.data)*(*s2).GetBitRate() != len(*(*s2).GetData())*s.bitRate {
		return ErrorNotEqualSoundDuration
	}
	s.authors = append(s.authors, (*s2).GetAuthors()...)
	if (*s2).GetBitRate() != s.bitRate {
		(*s2).RebitSound(s.bitRate) // change s2 bitrate to s bitrate
	}

	for i := 0; i < len(*s.data); i++ {
		(*s.data)[i] += (*(*s2).GetData())[i]
	}

	return nil
}

func (s *SoundWav) Append(s2 *sound.Sound) {
	if s.data == nil {
		return
	}

	if (*s2).GetBitRate() != s.bitRate { // TODO IT IS BAD SUTIATION AND NEED TO WARN THERE
		(*s2).RebitSound(s.bitRate) // change s2 bitrate to s bitrate
	}

	s.authors = append(s.authors, (*s2).GetAuthors()...)
	s.timeSend = append(s.timeSend, (*s2).GetTimeId()...)
	(*s.data) = append(*s.data, *(*s2).GetData()...)
	s.duration += (*s2).GetSoundDuration()
}

// partDuration in ms
// make TimeSend empty array
func (s *SoundWav) DivideIntoParts(partDuration int) ([]sound.Sound, error) { // TODO bad work with pointers
	if (s.duration*s.bitRate)%partDuration != 0 {
		return nil, ErrorIncorrectSoundDuration
	}

	amt := (s.duration * s.bitRate) / partDuration
	dividedSound := make([]sound.Sound, amt)
	for i := 0; i < (s.duration*s.bitRate)/partDuration; i++ {
		part := (*s.data)[i*s.bitRate*partDuration : (i+1)*s.bitRate*partDuration]
		dividedSound[i] = s.NewSound(&part, s.bitRate, s.duration, s.authors, make([]uint64, 0))
	}

	return dividedSound, nil
}

func (s *SoundWav) AmIAuthor(userId uint32) (bool, uint64) { // return 0 if not, otherwise timeSend // !!!!!!!!!CHECK IN PROTO THAT NO ONE SEND TIME 0
	for i, aId := range s.authors {
		if aId == userId {
			return true, s.timeSend[i]
		}
	}

	return false, 0
}

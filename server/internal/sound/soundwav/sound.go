package soundwav

import (
	"errors"
	"fmt"
	"homework/server/internal/sound"
)

type SoundWav struct { // all fieldS are public because we have to translate it to redis
	Data     *[]float32
	BitRate  int
	Duration int
	Authors  []uint32
	TimeSend []uint64
}

var ErrorNotEqualSoundDuration error = errors.New("incorrect sound wave duration - they must be equal")
var ErrorIncorrectSoundDuration error = errors.New("can't divide sound into parts as it has wrong duration or bitRate")
var ErrorIncorrectSoundType error = errors.New("incorrect sound type in divide into parts")
var ErrorCantRebitIncorrectBitrate error = errors.New("sound bitrate is probably not dividable by new bitrate or vice versa")

func NewSound(data *[]float32, bitRate int, duration int, authors []uint32, timeSend []uint64) *SoundWav {
	return &SoundWav{Data: data, BitRate: bitRate, Duration: duration, Authors: authors, TimeSend: timeSend}
}

func NewEmptySound() sound.Sound {
	return &SoundWav{}
}

func (s *SoundWav) NewEmptySound() sound.Sound {
	return &SoundWav{Data: nil, BitRate: 0, Duration: 0, Authors: []uint32{}, TimeSend: []uint64{}}
}

func (s *SoundWav) NewSound(data *[]float32, bitRate int, duration int, authors []uint32, timeSend []uint64) sound.Sound {
	return &SoundWav{Data: nil, BitRate: 0, Duration: 0, Authors: []uint32{}, TimeSend: []uint64{}}
}

func (s *SoundWav) String() string {
	return fmt.Sprintf("BitRate: %d, Duration: %d, Authors: %v, TimeSend: %v", s.BitRate, s.Duration, s.Authors, s.TimeSend)
}

func (s *SoundWav) GetData() *[]float32 {
	return s.Data
}

func (s *SoundWav) GetAuthors() []uint32 {
	return s.Authors
}

func (s *SoundWav) GetSoundDuration() int { // in nerly to ms because 1024 not 1000
	return 1024 * len(*s.Data) / s.BitRate
}

func (s *SoundWav) GetBitRate() int {
	return s.BitRate
}

func (s *SoundWav) GetTimeId() []uint64 {
	return s.TimeSend
}

func (s *SoundWav) SetTimeId(id uint64) {
	s.TimeSend = append(s.TimeSend, id)
}

// TODO check what is better - to use avarage or just take every kth
func (s *SoundWav) RebitSound(newBitRate int) error {
	if s.BitRate > newBitRate {
		if s.BitRate%newBitRate != 0 {
			return ErrorCantRebitIncorrectBitrate
		}

		k := s.BitRate / newBitRate
		for i := 0; i*k < len(*s.Data); i++ {
			(*s.Data)[i] = (*s.Data)[i*k]
		}

		(*s.Data) = (*s.Data)[:len(*s.Data)/k]
	} else {
		if newBitRate%s.BitRate != 0 {
			return ErrorCantRebitIncorrectBitrate
		}

		// it is very bad for sound quality if I understand correctly
		k := newBitRate / s.BitRate
		(*s.Data) = append(*s.Data, make([]float32, len(*s.Data)*(k-1))...)
		for i := len(*s.Data)/k - 1; i >= 0; i-- {
			for j := 0; j < k; j++ {
				(*s.Data)[i*k+j] = (*s.Data)[i]
			}
		}
	}

	return nil
}

func (s *SoundWav) Add(s2 *sound.Sound) error {
	if s == nil {
		return errors.New("s is nil")
	}
	if s2 == nil {
		return errors.New("s2 is nil")
	}
	if s.Data == nil {
		return errors.New("s.data is nil")
	}
	if (*s2).GetData() == nil {
		return errors.New("s2.data is nil")
	}

	if len(*s.Data)*(*s2).GetBitRate() != len(*(*s2).GetData())*s.BitRate { // TODO Check for cotecteness
		return ErrorNotEqualSoundDuration
	}

	s.Authors = append(s.Authors, (*s2).GetAuthors()...)
	if (*s2).GetBitRate() != s.BitRate {
		if err := (*s2).RebitSound(s.BitRate); err != nil { // change s2 bitrate to s bitrate
			return err
		}
	}

	if s == nil {
		return errors.New("s is nil")
	}
	if s2 == nil {
		return errors.New("s2 is nil")
	}
	if s.Data == nil {
		return errors.New("s.data is nil")
	}
	if (*s2).GetData() == nil {
		return errors.New("s2.data is nil")
	}

	for i := 0; i < len(*s.Data); i++ {
		(*s.Data)[i] += (*(*s2).GetData())[i]
	}

	return nil
}

func (s *SoundWav) Append(s2 *sound.Sound) {
	if s.Data == nil {
		return
	}

	if (*s2).GetBitRate() != s.BitRate {
		//s.logger.Warn("Have to rebit sound during append", zap.Int("Server bitRate", s.BitRate), zap.Int("User bitRate", (*s2).GetBitRate()))
		(*s2).RebitSound(s.BitRate) // change s2 bitrate to s bitrate
	}

	s.Authors = append(s.Authors, (*s2).GetAuthors()...)
	s.TimeSend = append(s.TimeSend, (*s2).GetTimeId()...)
	(*s.Data) = append(*s.Data, *(*s2).GetData()...)
	s.Duration += (*s2).GetSoundDuration()
}

// partDuration in ms
// make TimeSend empty array
func (s *SoundWav) DivideIntoParts(partDuration int) ([]sound.Sound, error) { // TODO bad work with pointers
	if (s.Duration*s.BitRate)%partDuration != 0 {
		return nil, ErrorIncorrectSoundDuration
	}

	amt := (s.Duration * s.BitRate) / partDuration
	dividedSound := make([]sound.Sound, amt)
	for i := 0; i < (s.Duration*s.BitRate)/partDuration; i++ {
		part := (*s.Data)[i*s.BitRate*partDuration : (i+1)*s.BitRate*partDuration]
		dividedSound[i] = s.NewSound(&part, s.BitRate, s.Duration, s.Authors, make([]uint64, 0))
	}

	return dividedSound, nil
}

func (s *SoundWav) AmIAuthor(userId uint32) (bool, uint64) { // return 0 if not, otherwise timeSend // !!!!!!!!!CHECK IN PROTO THAT NO ONE SEND TIME 0
	for i, aId := range s.Authors {
		if aId == userId {
			return true, s.TimeSend[i]
		}
	}

	return false, 0
}

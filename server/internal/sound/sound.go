package sound

type Sound interface {
	NewEmptySound() Sound
	NewSound(*[]float32, int, int, []uint32, []uint64) Sound
	String() string
	GetData() *[]float32
	GetAuthors() []uint32
	GetSoundDuration() int
	GetTimeId() []uint64
	SetTimeId(uint64)
	RebitSound(int) error
	GetBitRate() int
	Add(*Sound) error
	Append(*Sound)
	AmIAuthor(uint32) (bool, uint64)
	DivideIntoParts(partDuration int) ([]Sound, error)
}

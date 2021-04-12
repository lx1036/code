package bitmask

// BitMask interface allows hint providers to create BitMasks for TopologyHints
type BitMask interface {
	Add(bits ...int) error
	Remove(bits ...int) error
	And(masks ...BitMask)
	Or(masks ...BitMask)
	Clear()
	Fill()
	IsEqual(mask BitMask) bool
	IsEmpty() bool
	IsSet(bit int) bool
	AnySet(bits []int) bool
	IsNarrowerThan(mask BitMask) bool
	String() string
	Count() int
	GetBits() []int
}

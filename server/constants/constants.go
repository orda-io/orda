package constants

import "math"

// InfinitySseq is infinite number of sseq
const InfinitySseq uint64 = math.MaxUint64

const (
	TagServer       = "SERV"
	TagReset        = "CORS"
	TagCreate       = "COCR"
	TagClient       = "CLIE"
	TagPushPull     = "PUPU"
	TagPostPushPull = "POPP"
	TagTest         = "TEST"
)

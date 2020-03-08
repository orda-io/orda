package model

import (
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/stretchr/testify/require"
	"math"
	"testing"
)

func TestCheckType(t *testing.T) {
	i64Max, err := ConvertType(math.MaxInt64) // 9223372036854775807
	require.NoError(t, err)
	log.Logger.Infof("[%v]", i64Max)

	i64Min, err := ConvertType(math.MinInt64) // -9223372036854775808
	require.NoError(t, err)
	log.Logger.Infof("[%v]", i64Min)

	u64Max, err := ConvertType(uint64(math.MaxUint64)) // 18446744073709551615
	require.NoError(t, err)
	log.Logger.Infof("[%v]", u64Max)

	f64Max, err := ConvertType(math.MaxFloat64) // 1.7976931348623157e+308
	require.NoError(t, err)
	log.Logger.Infof("[%v]", f64Max)

	str, err := ConvertType("hello, world")
	require.NoError(t, err)
	log.Logger.Infof("[%v]", str)

	boolStr, err := ConvertType(true)
	require.NoError(t, err)
	log.Logger.Infof("[%v]", boolStr)

	strt := &struct {
		B bool
		S string
		A []int
	}{
		true,
		"hello world",
		[]int{1, 2, 3},
	}
	strtStr, err := ConvertType(strt)
	log.Logger.Infof("[%v]", strtStr)
}

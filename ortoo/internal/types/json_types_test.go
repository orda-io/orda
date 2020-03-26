package types

import (
	"github.com/knowhunger/ortoo/ortoo/log"
	"github.com/stretchr/testify/require"
	"math"
	"reflect"
	"testing"
)

func TestCheckType(t *testing.T) {
	i64Max := ConvertToJSONSupportedType(math.MaxInt64) // 9223372036854775807
	log.Logger.Infof("%v", reflect.TypeOf(i64Max))
	require.Equal(t, int64(math.MaxInt64), i64Max)
	log.Logger.Infof("[%v vs. %v]", i64Max, math.MaxInt64)

	i64Min := ConvertToJSONSupportedType(math.MinInt64) // -9223372036854775808
	require.Equal(t, int64(math.MinInt64), i64Min)
	log.Logger.Infof("[%v vs. %v]", i64Min, math.MinInt64)

	u64Max := ConvertToJSONSupportedType(uint64(math.MaxUint64)) // 18446744073709551615
	require.Equal(t, int64(-1), u64Max)

	log.Logger.Infof("%v %v", reflect.TypeOf(u64Max), u64Max)

	f64Max := ConvertToJSONSupportedType(math.MaxFloat64) // 1.7976931348623157e+308
	require.Equal(t, math.MaxFloat64, f64Max)
	log.Logger.Infof("[%v vs. %v]", f64Max, math.MaxFloat64)

	str := ConvertToJSONSupportedType("hello, world")
	require.Equal(t, "hello, world", str)
	log.Logger.Infof("[%v]", str)

	b := ConvertToJSONSupportedType(true)
	require.Equal(t, true, b)
	log.Logger.Infof("[%v]", b)

	strt := &struct {
		B bool
		S string
		A []int
	}{
		true,
		"hello world",
		[]int{1, 2, 3},
	}
	strtStr := ConvertToJSONSupportedType(strt)
	require.Equal(t, strt, strtStr)
	log.Logger.Infof("[%v]", strtStr)
}

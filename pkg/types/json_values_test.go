package types

import (
	"math"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/orda-io/orda/pkg/log"
)

func TestConvertToJSONSupportedType(t *testing.T) {
	t.Run("Can convert JSON supported types", func(t *testing.T) {
		i64Max := ConvertToJSONSupportedValue(math.MaxInt64) // 9223372036854775807
		log.Logger.Infof("%v", reflect.TypeOf(i64Max))
		require.Equal(t, float64(math.MaxInt64), i64Max)
		log.Logger.Infof("[%v vs. %v]", i64Max, math.MaxInt64)

		i64Min := ConvertToJSONSupportedValue(math.MinInt64) // -9223372036854775808
		require.Equal(t, float64(math.MinInt64), i64Min)
		log.Logger.Infof("[%v vs. %v]", i64Min, math.MinInt64)

		u64Max := ConvertToJSONSupportedValue(uint64(math.MaxUint64)) // 18446744073709551615
		require.Equal(t, float64(math.MaxUint64), u64Max)

		log.Logger.Infof("%v %v", reflect.TypeOf(u64Max), u64Max)

		f64Max := ConvertToJSONSupportedValue(math.MaxFloat64) // 1.7976931348623157e+308
		require.Equal(t, math.MaxFloat64, f64Max)
		log.Logger.Infof("[%v vs. %v]", f64Max, math.MaxFloat64)

		str := ConvertToJSONSupportedValue("hello, world")
		require.Equal(t, "hello, world", str)
		log.Logger.Infof("[%v]", str)

		b := ConvertToJSONSupportedValue(true)
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
		strtStr := ConvertToJSONSupportedValue(strt)
		require.Equal(t, strt, strtStr)
		log.Logger.Infof("[%v]", strtStr)
	})

	t.Run("Can convert JSON supported pointer types", func(t *testing.T) {
		var intt = 1234
		cintt := ConvertToJSONSupportedValue(&intt)
		require.Equal(t, cintt, float64(intt))
	})

}

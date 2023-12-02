package utils

import (
	"fmt"
	"math"
	"strconv"

	"github.com/cockroachdb/apd"
	"github.com/pkg/errors"
)

var apdContext *apd.Context = apd.BaseContext.WithPrecision(20) // TODO: move to config
var Zero *apd.Decimal = Reduce(apd.New(0, 0))                   // Shoudln't be mutated
var One *apd.Decimal = Reduce(apd.New(1, 0))                    // Shoudln't be mutated
var D10 *apd.Decimal = Reduce(apd.New(10, 0))                   // Shoudln't be mutated
var D100 *apd.Decimal = Reduce(apd.New(100, 0))                 // Shoudln't be mutated

func Reduce(x *apd.Decimal) *apd.Decimal {
	result := NewZero()
	result.Reduce(x)
	return result
}

func NewZero() *apd.Decimal {
	return apd.New(0, 0)
}

func FromUint(x uint) *apd.Decimal {
	return Reduce(apd.New(int64(x), 0))
}

// Only 15 places precision
func FromFloat64(x float64) *apd.Decimal {
	xstring := strconv.FormatFloat(x, 'E', 15, 64)
	result, _, err := apdContext.NewFromString(xstring)
	if err != nil {
		panic(err)
	}
	return Reduce(result)
}

// Convert float64 to decimal
func SetFloat64(a float64) *apd.Decimal {
	result := NewZero()
	_, err := result.SetFloat64(a)
	if err != nil {
		panic(err)
	}
	return Reduce(result)
}

func FromString(x string) *apd.Decimal {
	result, _, err := apd.NewFromString(x)
	if err != nil {
		panic(err)
	}
	return Reduce(result)
}

func FromStringErr(x string) (*apd.Decimal, error) {
	result, _, err := apd.NewFromString(x)
	if err != nil {
		return nil, err
	}
	return Reduce(result), nil
}

// a - b
func Sub(a, b *apd.Decimal) *apd.Decimal {
	result := NewZero()
	_, err := apdContext.Sub(result, a, b)
	if err != nil {
		panic(err) // TODO:
	}
	return Reduce(result)
}

// a + b
func Add(a, b *apd.Decimal) *apd.Decimal {
	result := NewZero()
	_, err := apdContext.Add(result, a, b)
	if err != nil {
		panic(err) // TODO:
	}
	return Reduce(result)
}

// a/b
func Mul(a, b *apd.Decimal) *apd.Decimal {
	result := NewZero()
	_, err := apdContext.Mul(result, a, b)
	if err != nil {
		panic(err) // TODO:
	}
	return Reduce(result)
}

// a/b
func Div(a, b *apd.Decimal) *apd.Decimal {
	result := NewZero()
	_, err := apdContext.Quo(result, a, b)
	if err != nil {
		panic(err) // TODO:
	}
	return Reduce(result)
}

// a%b
func Mod(a, b *apd.Decimal) *apd.Decimal {
	result := NewZero()
	_, err := apdContext.Rem(result, a, b)
	if err != nil {
		panic(err) // TODO:
	}
	return Reduce(result)
}

// |a|
func Abs(a *apd.Decimal) *apd.Decimal {
	result := NewZero()
	_, err := apdContext.Abs(result, a)
	if err != nil {
		panic(err) // TODO:
	}
	return result
}

// -a
func Neg(a *apd.Decimal) *apd.Decimal {
	result := NewZero()
	_, err := apdContext.Neg(result, a)
	if err != nil {
		panic(err) // TODO:
	}
	return result
}

func Round(a *apd.Decimal) *apd.Decimal {
	result := NewZero()
	_, err := apd.BaseContext.RoundToIntegralValue(result, a)
	if err != nil {
		panic(err) // TODO:
	}
	return result
}

const MinPreciseFloat64Integer int64 = -(1 << 53)
const MaxPreciseFloat64Integer int64 = 1 << 53

func ToIntegerInFloat64(a *apd.Decimal) (float64, error) {
	num, err := a.Int64()
	if err != nil {
		return 0, errors.Wrap(err, "cannot be represented with int64")
	}
	if num < MinPreciseFloat64Integer || MaxPreciseFloat64Integer < num {
		return 0, errors.New("cannot be represented as integers in float64")
	}
	return float64(num), nil
}

// a < b
func Less(a, b *apd.Decimal) bool {
	return a.Cmp(b) < 0
}

// a <= b
func LessOrEq(a, b *apd.Decimal) bool {
	return a.Cmp(b) <= 0
}

// a > b
func Greater(a, b *apd.Decimal) bool {
	return a.Cmp(b) > 0
}

// a >= b
func GreaterOrEq(a, b *apd.Decimal) bool {
	return a.Cmp(b) >= 0
}

// a == b
func Eq(a, b *apd.Decimal) bool {
	return a.Cmp(b) == 0
}

// Returns flat, non scientific string representation
// 1000 will be "1000" not like this "1E3"
// 0.001 will be as "0.001" not like "1E-3"
func ToFlatString(x *apd.Decimal) string {
	return fmt.Sprintf("%f", x)
}

func RoundFloat(val float64, precision uint32) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func FloorFloat(val float64, precision uint32) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Floor(val*ratio) / ratio
}

func CeilFloat(val float64, precision uint32) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Ceil(val*ratio) / ratio
}

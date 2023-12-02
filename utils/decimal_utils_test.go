package utils

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"

	"github.com/cockroachdb/apd"
	"github.com/stretchr/testify/assert"
)

func TestReduce(t *testing.T) {
	x1, _, _ := apd.NewFromString("100")
	assert.Equal(t, "100", x1.String())
	assert.Equal(t, "1E+2", Reduce(x1).String())

	x2, _, _ := apd.NewFromString("100.0")
	assert.Equal(t, "100.0", x2.String())
	assert.Equal(t, "1E+2", Reduce(x2).String())

	x3, _, _ := apd.NewFromString("100.00")
	assert.Equal(t, "100.00", x3.String())
	assert.Equal(t, "1E+2", Reduce(x3).String())
}

func TestConstants(t *testing.T) {
	assert.Equal(t, "0", Zero.String())
	assert.Equal(t, "1", One.String())
	assert.Equal(t, "1E+2", D100.String())
}

func TestZero(t *testing.T) {
	assert.Equal(t, "0", NewZero().String())
	assert.NotSame(t, NewZero(), NewZero())
}

func TestFromUint(t *testing.T) {
	x150, err := FromUint(150).Float64()
	assert.Nil(t, err)
	assert.Equal(t, 150.0, x150)

	x0, err := FromUint(0).Float64()
	assert.Nil(t, err)
	assert.Equal(t, 0.0, x0)

	x123456789, err := FromUint(123456789).Float64()
	assert.Nil(t, err)
	assert.Equal(t, 123456789.0, x123456789)
}

func TestFromFloat(t *testing.T) {
	x0pt3 := FromFloat64(0.3).String()
	assert.Equal(t, "0.3", x0pt3)
	assert.Equal(t, "0.299999999999999989", strconv.FormatFloat(0.3, 'f', 18, 64))
}

func TestFromString(t *testing.T) {
	str := "123123124234345324524234234.14567891234852375837452745897138957195"
	decimal := FromString(str)
	assert.Equal(t, str, decimal.String())
}

func TestFromStringErr(t *testing.T) {
	str := "123123124234345324524234234.14567891234852375837452745897138957195"
	decimal, err := FromStringErr(str)
	assert.NoError(t, err)
	assert.Equal(t, str, decimal.String())

	_, err = FromStringErr("Wrong")
	assert.Error(t, err)
}

func TestSub(t *testing.T) {
	x15 := FromUint(15)
	x23 := FromUint(23)
	assert.Equal(t, "-8", Sub(x15, x23).String())

	x0_55, _, _ := apd.NewFromString("0.55")
	x0_15, _, _ := apd.NewFromString("0.15")
	assert.Equal(t, "0.4", Sub(x0_55, x0_15).String())
}

func TestAdd(t *testing.T) {
	x15 := FromUint(15)
	x23 := FromUint(23)
	assert.Equal(t, "38", Add(x15, x23).String())

	x0_55, _, _ := apd.NewFromString("0.55")
	x0_15, _, _ := apd.NewFromString("0.15")
	assert.Equal(t, "0.7", Add(x0_55, x0_15).String())
}

func TestMul(t *testing.T) {
	x5 := FromUint(5)
	x10 := FromUint(10)
	assert.Equal(t, "50", fmt.Sprintf("%f", Mul(x5, x10)))

	x0_15, _, _ := apd.NewFromString("0.15")
	assert.Equal(t, "1.5", fmt.Sprintf("%f", Mul(x10, x0_15)))
}

func TestDiv(t *testing.T) {
	x5 := FromUint(5)
	x10 := FromUint(10)
	assert.Equal(t, "0.5", fmt.Sprintf("%f", Div(x5, x10)))
}

func TestAbs(t *testing.T) {
	x17 := Abs(FromUint(17))
	assert.Equal(t, "17", x17.String())

	xneg17 := Abs(FromFloat64(-17))
	assert.Equal(t, "17", xneg17.String())
}

func TestNeg(t *testing.T) {
	x17 := Neg(FromUint(17))
	assert.Equal(t, "-17", x17.String())

	xneg17 := Neg(FromFloat64(-17))
	assert.Equal(t, "17", xneg17.String())
}

func TestRound(t *testing.T) {
	x1p6 := Round(FromString("1.6"))
	assert.Equal(t, "2", x1p6.String())

	x1p5 := Round(FromString("1.5"))
	assert.Equal(t, "2", x1p5.String())

	x1p4 := Round(FromString("1.4"))
	assert.Equal(t, "1", x1p4.String())

	xm0p8 := Round(FromString("-0.8"))
	assert.Equal(t, "-1", xm0p8.String())

	xm0p5 := Round(FromString("-0.5"))
	assert.Equal(t, "-1", xm0p5.String())

	xm0p4 := Round(FromString("-0.4"))
	assert.True(t, Eq(xm0p4, Zero))
}

func TestLess(t *testing.T) {
	x17 := FromUint(17)
	y17 := FromUint(17)
	assert.False(t, Less(x17, y17))
	x18 := FromUint(18)
	assert.True(t, Less(x17, x18))
	assert.False(t, Less(x18, x17))
}

func TestLessOrEq(t *testing.T) {
	x17 := FromUint(17)
	y17 := FromUint(17)
	assert.True(t, LessOrEq(x17, y17))
	x18 := FromUint(18)
	assert.True(t, LessOrEq(x17, x18))
	assert.False(t, LessOrEq(x18, x17))
}

func TestGreater(t *testing.T) {
	x17 := FromUint(17)
	y17 := FromUint(17)
	assert.False(t, Greater(x17, y17))
	x18 := FromUint(18)
	assert.True(t, Greater(x18, x17))
	assert.False(t, Greater(x17, x18))
}

func TestGreaterOrEq(t *testing.T) {
	x17 := FromUint(17)
	y17 := FromUint(17)
	assert.True(t, GreaterOrEq(x17, y17))
	x18 := FromUint(18)
	assert.True(t, GreaterOrEq(x18, x17))
	assert.False(t, GreaterOrEq(x17, x18))
}

func TestEq(t *testing.T) {
	x17 := apd.New(170, 1)
	y17 := apd.New(17, 2)
	assert.True(t, Eq(x17, y17))
	x18 := FromUint(18)
	assert.False(t, Eq(x18, x17))
	assert.False(t, Eq(x17, x18))
}

func TestToFlatString(t *testing.T) {
	x10000_1 := apd.New(1, 4)
	assert.Equal(t, "10000", ToFlatString(x10000_1))

	x10000_2 := apd.New(10, 3)
	assert.Equal(t, "10000", ToFlatString(x10000_2))

	x0_0001_1 := apd.New(1, -4)
	assert.Equal(t, "0.0001", ToFlatString(x0_0001_1))

	x0_0001_2 := apd.New(10, -5)
	assert.Equal(t, "0.00010", ToFlatString(x0_0001_2))

	x1000zeroes := apd.New(1, 1000)
	buf := &bytes.Buffer{}
	buf.WriteString("1")
	for i := 0; i < 1000; i++ {
		buf.WriteRune('0')
	}
	assert.Equal(t, buf.String(), ToFlatString(x1000zeroes))
}

func TestToIntegerInFloat64(t *testing.T) {
	x, err := ToIntegerInFloat64(FromString("1"))
	assert.NoError(t, err)
	assert.Equal(t, 1.0, x)

	x, err = ToIntegerInFloat64(FromString("0"))
	assert.NoError(t, err)
	assert.Equal(t, 0.0, x)

	x, err = ToIntegerInFloat64(FromString("-1"))
	assert.NoError(t, err)
	assert.Equal(t, -1.0, x)

	_, err = ToIntegerInFloat64(FromString("1e100"))
	assert.Error(t, err)

	_, err = ToIntegerInFloat64(FromString("-1e100"))
	assert.Error(t, err)

	_, err = ToIntegerInFloat64(FromUint(1 << 62))
	assert.Error(t, err)

	_, err = ToIntegerInFloat64(Neg(FromUint(1 << 62)))
	assert.Error(t, err)

	_, err = ToIntegerInFloat64(FromString("0.5"))
	assert.Error(t, err)

	_, err = ToIntegerInFloat64(FromString("-0.5"))
	assert.Error(t, err)
}

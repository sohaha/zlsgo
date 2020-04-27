package zvalid

import (
	"testing"

	"github.com/sohaha/zlsgo"
)

func TestRuleHas(t *testing.T) {
	var err error
	tt := zlsgo.NewTest(t)

	err = New().Verifi("123a").HasLetter().Error()
	tt.EqualNil(err)
	err = New().Verifi("1").HasLetter().Error()
	tt.Equal(true, err != nil)
	t.Log(err)
	err = New().Verifi("").HasLetter().Error()
	tt.Equal(true, err != nil)

	err = New().Verifi("123a").HasLower().Error()
	tt.EqualNil(err)
	err = New().Verifi("1").HasLower().Error()
	tt.Equal(true, err != nil)
	err = New().Verifi("").HasLower().Error()
	tt.Equal(true, err != nil)

	err = New().Verifi("123aA").HasUpper().Error()
	tt.EqualNil(err)
	err = New().Verifi("1").HasUpper().Error()
	tt.Equal(true, err != nil)
	err = New().Verifi("").HasUpper().Error()
	tt.Equal(true, err != nil)

	err = New().Verifi("123aA").HasNumber().Error()
	tt.EqualNil(err)
	err = New().Verifi("a").HasNumber().Error()
	tt.Equal(true, err != nil)
	err = New().Verifi("").HasNumber().Error()
	tt.Equal(true, err != nil)

	err = New().Verifi("123aA.").HasSymbol().Error()
	tt.EqualNil(err)
	err = New().Verifi("a").HasSymbol().Error()
	tt.Equal(true, err != nil)
	err = New().Verifi("").HasSymbol().Error()
	tt.Equal(true, err != nil)

	err = New().Verifi("123aA.").HasString("aA").Error()
	tt.EqualNil(err)
	err = New().Verifi("a").HasString("c").Error()
	tt.Equal(true, err != nil)
	err = New().Verifi("").HasString("a").Error()
	tt.Equal(true, err != nil)

	err = New().Verifi("123aA.").HasPrefix("123a").Error()
	tt.EqualNil(err)
	err = New().Verifi("a").HasPrefix("c").Error()
	tt.Equal(true, err != nil)
	err = New().Verifi("").HasPrefix("a").Error()
	tt.Equal(true, err != nil)

	err = New().Verifi("123aA.").HasSuffix("A.").Error()
	tt.EqualNil(err)
	err = New().Verifi("a").HasSuffix("c").Error()
	tt.Equal(true, err != nil)
	err = New().Verifi("").HasSuffix("a").Error()
	tt.Equal(true, err != nil)

	err = New().Verifi("123aA.").Password().Error()
	tt.EqualNil(err)
	err = New().Verifi("a", "pass2").Password().Error()
	tt.Equal(true, err != nil)
	tt.Log(err)
	err = New().Verifi("").Password().Error()
	tt.Equal(true, err != nil)
	tt.Log(err)

	err = New().Verifi("123aA.").StrongPassword().Error()
	tt.EqualNil(err)
	err = New().Verifi("123aA").StrongPassword().Error()
	tt.Equal(true, err != nil)
	err = New().Verifi("").StrongPassword().Error()
	tt.Equal(true, err != nil)

}

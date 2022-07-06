package zstring_test

import (
	"fmt"
	"testing"

	zls "github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zstring"
)

func TestRSA(t *testing.T) {
	tt := zls.NewTest(t)

	val := "是我呀，我是测试的人呢，你想干嘛呀？？？我就是试试看这么长会发生什么情况呢"

	prv, pub, err := zstring.GenRSAKey()
	tt.NoError(err)

	fmt.Println(string(prv))
	fmt.Println(string(pub))

	c, err := zstring.RSAEncryptString(val, string(pub))
	tt.EqualNil(err)
	t.Log(c)

	c, err = zstring.RSADecryptString(c, string(prv))
	tt.EqualNil(err)
	t.Log(c)

	tt.Equal(val, c)

	c, err = zstring.RSAEncryptString(val, "pub")
	t.Log(c, err)

	c, err = zstring.RSADecryptString(c, "prv")
	t.Log(c, err)
}

func TestRSALong(t *testing.T) {
	tt := zls.NewTest(t)

	val := `是我呀，我是测试的人呢，你想干嘛呀？？？xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx 我就是试试看这么长会发生什么情况呢!
看起来似乎还行🤣🤣🤣!!!是我呀，我是测试的人呢，你想干嘛呀？？？xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx 我就是试试看这么长会发生什么情况呢!
看起来似乎还行🤣🤣🤣!!!是我呀，我是测试的人呢，你想干嘛呀？？？xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx 我就是试试看这么长会发生什么情况呢!
看起来似乎还行🤣🤣🤣!!!是我呀，我是测试的人呢，你想干嘛呀？？？xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx 我就是试试看这么长会发生什么情况呢!
看起来似乎还行🤣🤣🤣!!!是我呀，我是测试的人呢，你想干嘛呀？？？xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx 我就是试试看这么长会发生什么情况呢!
看起来似乎还行🤣🤣🤣!!!是我呀，我是测试的人呢，你想干嘛呀？？？xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx 我就是试试看这么长会发生什么情况呢!
看起来似乎还行🤣🤣🤣!!!
`

	prv, pub, err := zstring.GenRSAKey(2000)
	tt.NoError(err)

	b, err := zstring.RSAEncrypt([]byte(val), string(pub), 2000)
	t.Log(string(b), err)

	b, err = zstring.RSADecrypt(b, string(prv), 2000)
	t.Log(string(b), err)

	tt.Equal(val, string(b))

	b, err = zstring.RSAEncrypt([]byte("val2"), string(pub), 2000)
	t.Log(string(b), err)

}

func TestRSAPrvKeyEncrypt(t *testing.T) {
	tt := zls.NewTest(t)

	val := "是我呀，我是测试的人呢，你想干嘛呀？？？我就是试试看这么长会发生什么情况呢"

	prv, pub, err := zstring.GenRSAKey()
	tt.NoError(err)

	c, err := zstring.RSAPriKeyEncryptString(val, string(prv))
	tt.EqualNil(err)
	t.Log(c)

	c, err = zstring.RSAPubKeyDecryptString(c, string(pub))
	tt.EqualNil(err)
	t.Log(c)

	tt.Equal(val, c)

	c, err = zstring.RSAPriKeyEncryptString(val, "pub")
	t.Log(c, err)

	c, err = zstring.RSAPubKeyDecryptString(c, "prv")
	t.Log(c, err)
}

package zstring_test

import (
	zls "github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zstring"
	"testing"
)

const pub = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAtS8G5Ug3P2ra4MLyy+QL
0vaBT0OX96YiB71e0oiZ7QXreftbWmfPUEIlJ6E9dKlAIrS3JSbgWxZhqxRxVSql
DoGEPMWGjhECEalISuLH3Y0LiEXuQHmKlJgmlxpvg4b3+xbxCWLPZCnPdbj41AW1
peN2jJEiO84+SwgKM/0/6G8BPzY3ECAEk1BxCA9MbslXEwBs8N0uG8p8Uu1Jc/mu
503Z+zjhbFRfi7s2QGt8wadqh10q7YnDZ958Iqpk0nrIe1s7IzLB4618z2F/ZO3t
nub2vy2Hr2uhuldbpc3/v1ywVWiicdHFNW1hec2iLJ4e5Go9aF0egcOis0I3Imq7
KwIDAQAB
-----END PUBLIC KEY-----`

const pri = `-----BEGIN RSA PRIVATE KEY-----
MIIEogIBAAKCAQEAtS8G5Ug3P2ra4MLyy+QL0vaBT0OX96YiB71e0oiZ7QXreftb
WmfPUEIlJ6E9dKlAIrS3JSbgWxZhqxRxVSqlDoGEPMWGjhECEalISuLH3Y0LiEXu
QHmKlJgmlxpvg4b3+xbxCWLPZCnPdbj41AW1peN2jJEiO84+SwgKM/0/6G8BPzY3
ECAEk1BxCA9MbslXEwBs8N0uG8p8Uu1Jc/mu503Z+zjhbFRfi7s2QGt8wadqh10q
7YnDZ958Iqpk0nrIe1s7IzLB4618z2F/ZO3tnub2vy2Hr2uhuldbpc3/v1ywVWii
cdHFNW1hec2iLJ4e5Go9aF0egcOis0I3Imq7KwIDAQABAoIBAFv9Rmj+414FaJ+Z
GyC95er0UO7niK6p4LlBQnVt+YjH6qiCH/2kmzNKgga+7K7gh7mXOy1Xsa1NjcUI
mgn9ntPgmj0opIpYxE4nPpcW0RcBV4uWxcJicyPCpEUvnNKQojMPkM2NJ3LZb4V7
poovY+yXskboIRNwQVxi9psyx1HAvy4SJ6884W6IBxgVQ3M0k1w2mnLnqOKmxYcD
KSIPwGpevYUygTiQQFoy9BKQYvJHnK9bhdTtxiSdC1FFbjE4sao7T8UxrwYmybyq
pTLTycUIwbHr9it/UdOAs6rcbBUT5xF4/Gen9jATk34z8BYbhGbXVjlrC4dtxwU/
XH50LsECgYEA5OR4eq2g+DJ67jeLFqtUCzQb5vvtbZB+vRbk3z/l/FNYalqhO0ne
seXxdOETPDUe3dlU5USqgGmqEtWYce1iD3f5og043SWy2PcgOYdXtRU3OWP6wBfG
CHlYsJ6G6Bvl8GFSEi1ePR+U9sMvdRZHb1FudKZt47rdkIBeGYr8NRsCgYEAyqQg
9KVN0aYse8A37EUM8zu1rdBcIIss6Mg42m38/1wvmemPJuNYPDVTt3OHipaD+Nh/
+hLFawM6LqBjFzRCBAEUvOtPpVEJrKKYoM8hZSuam2/thTMvwHhS47/s+f2qNEGU
ICZhqwOwqL3atTgaAmrHPlOon/3WR7lANgsjwzECgYB/EQ6JHCaGYo+3+wGt3gLU
DWOIAUc3UcXp9vGrte9Y+nPU5ucm4MVOARbgCasB+4NdKS9l746vpvkRZ54vcNbF
O5dLjQeKTUlSBS7QgQABuPtlUsl7Jjd7sNG5iufdps8peP10tdbhG804h/aqi2mw
tIYbH+FVUQF7HKggifWlDQKBgDQnJ8AvJycU8I/s+beaUenr7SdN39gUWbuThGZb
Nmj2bd3b6Zblnhjo1KH7XuABOvf5qH5RBHQ1QW0spDQdo/vp10+D9FykzaubsVJ5
3KtwHHtyxBuq/9g2X4b0J2ZzrbGDSz83AZ4E9huHuVk4liEXIC5fU5/RsauF9wux
tEORAoGACwZmskTmxW5TNWvjLCJc/sHdKlXqSIZ1FDG6v2xqMbAiRcsSc2FT4a6J
DLbJz0pCMufeXYFXHa2vTmwaWJTEuV/s68/s3cSCZd0RhU9DFYvAfygsc/uapYdH
chPduY1olk3Ry9KRHojSp2j8DtUGY7rWCnIfxWH4roLOcgyrxLA=
-----END RSA PRIVATE KEY-----
`

func TestRSA(t *testing.T) {
	tt := zls.NewTest(t)

	val := "是我呀，我是测试的人呢，你想干嘛呀？？？我就是试试看这么长会发生什么情况呢"

	c, err := zstring.RSAEncryptString(val, pub)
	tt.EqualNil(err)
	t.Log(c)

	c, err = zstring.RSADecryptString(c, pri)
	tt.EqualNil(err)
	t.Log(c)

	tt.Equal(val, c)

	c, err = zstring.RSAEncryptString(val, "pub")
	t.Log(c, err)

	c, err = zstring.RSADecryptString(c, "pri")
	t.Log(c, err)
}

func TestRSA2(t *testing.T) {
	tt := zls.NewTest(t)

	val := "是我呀，我是测试的人呢，你想干嘛呀？？？我就是试试看这么长会发生什么情况呢"

	c, err := zstring.RSAPriKeyEncryptString(val, pri)
	tt.EqualNil(err)
	t.Log(c)

	c, err = zstring.RSAPubKeyDecryptString(c, pub)
	tt.EqualNil(err)
	t.Log(c)

	tt.Equal(val, c)

	c, err = zstring.RSAPriKeyEncryptString(val, "pub")
	t.Log(c, err)

	c, err = zstring.RSAPubKeyDecryptString(c, "pri")
	t.Log(c, err)
}

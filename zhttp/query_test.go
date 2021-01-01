package zhttp_test

import (
	"strings"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zhttp"
)

const html = `
<html>
	<head>
		<title>Test</title>
	</head>
	<body>
ok
		<div id="Red" class="content red">is red box</div>
		<hr id="HR" />ha
		<br>
		<div class="content">is box</div>
		<div class="content test blue">blue box</div>
		<div class="content test blue blue2" data-name="TheSmurfs">blue2 box</div>
		<div class="multiple boxs">
			<div class="content">content:<span>666</span></div>
			<div id="Tow" name="saiya" class="content tow">div->div.tow<i>M</i></div>
		</div>
yes
	</body>
</html>
`

func TestHTMLParse(tt *testing.T) {
	t := zlsgo.NewTest(tt)

	h, err := zhttp.HTMLParse([]byte("<div></div>"))
	t.Log(h, err)

	h, err = zhttp.HTMLParse([]byte("div"))
	t.Log(h, err)

	h, err = zhttp.HTMLParse([]byte("<!- ok -->"))
	t.Log(h, err)

	h, err = zhttp.HTMLParse([]byte(""))
	t.Log(h.Attr("name"), err)

	h, err = zhttp.HTMLParse([]byte("<html><div class='box'>The is HTML</div><div class='red'>Red</div></html>"))
	t.Log(h.Select("div").Text(), h.Select("div").Attr("class"), h.Select("div", map[string]string{"class": "red"}).HTML())

	h, err = zhttp.HTMLParse([]byte(html))
	if err != nil {
		tt.Fatal(err)
	}

	t.Log(len(h.Select("body").FullText()))

	t.EqualExit("okhayes", strings.TrimSpace(strings.Replace(strings.Replace(h.Select("body").Text(), "\n", "", -1), "\t", "", -1)))

	// does not exist
	el := h.Select("div", map[string]string{"id": "does not exist"})
	tt.Logf("Not: %s|%s|%s\n", el.Attr("name"), el.Text(), el.HTML())

	// Tow
	el = h.Select("div", map[string]string{"ID": "Tow"})
	tt.Logf("Tow: %s|%s|%s\n", el.Attr("name"), el.Text(), el.HTML())

	// Tow
	el = h.Select("div", map[string]string{"Id": "Tow"})
	tt.Logf("Tow: %s|%s|%s\n", el.Attr("name"), el.Text(), el.HTML())

	// Blue
	el = h.Select("div", map[string]string{"class": "blue"})
	tt.Logf("Blue: %s|%s|%s\n", el.Attr("data-name"), el.Text(), el.HTML())

	// Blue2
	el = h.Select("div", map[string]string{"class": " blue blue2 "})
	tt.Logf("Blue2: %s|%s|%s\n", el.Attr("data-name"), el.Text(), el.HTML())

}

func TestSelectAll(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	h, _ := zhttp.HTMLParse([]byte(html))
	t.EqualExit(7, len(h.SelectAll("div")))
	t.EqualExit(0, len(h.SelectAll("vue")))
}

func TestSelectParent(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	h, _ := zhttp.HTMLParse([]byte(html))
	span := h.Select("span")
	tt.Log(span.Exist())

	multiple := span.SelectParent("div", map[string]string{"class": "multiple"})
	tt.Log(multiple.HTML())
	t.EqualTrue(multiple.Exist())

	parent := span.SelectParent("vue")
	tt.Log(parent.HTML())
	t.EqualTrue(!parent.Exist())
}

func TestChild(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	h, _ := zhttp.HTMLParse([]byte(html))

	child := h.Select("body").Child()
	for i, v := range child {
		switch i {
		case 1:
			t.EqualExit("hr", v.Name())
		case 2:
			t.EqualExit("br", v.Name())
		case 5:
			t.EqualExit("div", v.Name())
			t.EqualExit("blue2 box", v.Text())
		}
	}

	multiple := h.Select("body").Select("div", map[string]string{"class": "multiple"})

	span := multiple.SelectChild("span")
	tt.Log(span.Exist(), span.HTML())

	span = multiple.Select("div", map[string]string{"class": "content"}).SelectChild("span")
	t.EqualTrue(span.Exist())
	tt.Log(span.HTML())
}

func TestFind(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	h, _ := zhttp.HTMLParse([]byte(html))
	t.EqualTrue(h.Find("div#Tow.tow").Exist())
	t.EqualTrue(h.Find("div.multiple.boxs .tow>i").Exist())
	t.EqualTrue(!h.Find("div.multiple.boxs >   i").Exist())

	t.Log(h.Select("div", map[string]string{"class": "multiple boxs"}).Select("", map[string]string{"class": "tow"}).SelectChild("i").Exist())
	t.Log(h.Select("div", map[string]string{"class": "multiple boxs"}).Select("", map[string]string{"class": "tow"}).SelectChild("i").HTML())

}

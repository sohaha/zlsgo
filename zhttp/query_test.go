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
		<div class="content test blue">

blue box </div>
		<div class="content test blue blue2" data-name="TheSmurfs">blue2 box</div>
		<div class="multiple boxs">
			<div id="One" class="content">content:<span>666</span></div>
			<div id="Tow" name="saiya" class="content tow">div->div.tow<i>M</i></div>
			<div id="Three">Three</div>
			<div id="Four">Four</div>
		</div>
yes
	</body>
</html>
`

func TestHTMLParse(tt *testing.T) {
	t := zlsgo.NewTest(tt)

	h, err := zhttp.HTMLParse([]byte("<div></div>"))
	tt.Log(h, err)

	h, err = zhttp.HTMLParse([]byte("div"))
	tt.Log(h, err)

	h, err = zhttp.HTMLParse([]byte("<!- ok -->"))
	tt.Log(h, err)

	h, err = zhttp.HTMLParse([]byte(""))
	tt.Log(h.Attr("name"), err)

	h, err = zhttp.HTMLParse([]byte("<html><div class='box'>The is HTML</div><div class='red'>Red</div></html>"))
	tt.Log(h.Select("div").Text(), h.Select("div").Attr("class"),
		h.Select("div", map[string]string{"class": "red"}).HTML())

	h, err = zhttp.HTMLParse([]byte(html))
	if err != nil {
		tt.Fatal(err)
	}

	tt.Log(len(h.Select("body").FullText(true)))

	t.EqualExit("okhayes", strings.Replace(strings.Replace(h.Select("body").Text(), "\n", "", -1), "\t", "", -1))
	t.EqualExit("blue box", h.Find(".blue").Text(true))

	// does not exist
	el := h.Select("div", map[string]string{"id": "does not exist"})
	tt.Logf("id Not: %s|%s|%s\n", el.Attr("name"), el.Text(), el.HTML())

	// Tow
	el = h.Select("div", map[string]string{"id": "Tow"})
	tt.Logf("id Tow: %s|%s|%s\n", el.Attr("name"), el.Text(), el.HTML())
	t.EqualTrue(el.Exist())
	t.EqualExit("Tow", el.Attr("id"))

	// Tow
	el = h.Select("div", map[string]string{"Id": "Tow"})
	tt.Logf("Id Tow: %s|%s|%s\n", el.Attr("name"), el.Text(), el.HTML())

	// Blue
	el = h.Select("div", map[string]string{"class": "blue"})
	tt.Logf("class Blue: %s|%s|%s\n", el.Attr("data-name"), el.Text(true),
		el.HTML(true))

	// Blue2
	el = h.Select("div", map[string]string{"class": " blue blue2 "})
	tt.Logf("class Blue2: %s|%s|%s\n", el.Attr("data-name"), el.Text(true),
		el.HTML(true))

}

func TestSelectAll(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	h, _ := zhttp.HTMLParse([]byte(html))
	t.EqualExit(9, len(h.SelectAll("div")))
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

	t.Equal("", multiple.NthChild(0).Attr("id"))
	t.Equal("One", multiple.NthChild(1).Attr("id"))
	t.Equal("Tow", multiple.NthChild(2).Attr("id"))
	t.Equal("Three", multiple.NthChild(3).Attr("id"))

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

func TestBrother(tt *testing.T) {
	t := zlsgo.NewTest(tt)
	t.Log("oik")
	h, _ := zhttp.HTMLParse([]byte(html))
	d := h.Find("#Tow")
	parent := d.SelectParent("")
	tt.Log(parent.HTML())

	b := d.SelectBrother("div")
	t.EqualTrue(b.Exist())
	t.Equal("Three", b.Attr("id"))
	tt.Log(b.HTML(true))

	b = b.SelectBrother("div")
	t.EqualTrue(b.Exist())
	t.Equal("Four", b.Attr("id"))
	tt.Log(b.HTML(true))

	b = b.SelectBrother("div")
	t.EqualTrue(!b.Exist())

	b = h.Find("#One").SelectBrother("")
	tt.Log(b.HTML(true))
	t.EqualTrue(b.Exist())
}

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
		<div>
			<div class="content">div</div>
			<div id="Tow" name="saiya" class="content tow">div->div.tow</div>
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
	t.Log(h, err)

	h, err = zhttp.HTMLParse([]byte(html))
	if err != nil {
		tt.Fatal(err)
	}

	t.Log(len(h.Find("body").FullText()))

	t.EqualExit("okhayes", strings.TrimSpace(strings.Replace(strings.Replace(h.Find("body").Text(), "\n", "", -1), "\t", "", -1)))

	child := h.Find("body").Child()
	for i, v := range child {
		tt.Log(i, v.Name(), v.Text())
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

	t.Log(h.FindAll("div")[0].HTML())
	t.Log(h.MustFindAll("vue"))

	doc, err := h.MustFind("div")
	t.EqualTrue(err == nil)
	if err == nil {
		tt.Log(doc.Text())
	}

	// does not exist
	el := h.Find("div", map[string]string{"id": "does not exist"})
	tt.Logf("Not: %s|%s|%s\n", el.Attr("name"), el.Text(), el.HTML())

	// Tow
	el = h.Find("div", map[string]string{"id": "Tow"})
	tt.Logf("Tow: %s|%s|%s\n", el.Attr("name"), el.Text(), el.HTML())

	// Blue
	el = h.Find("div", map[string]string{"class": "blue"})
	tt.Logf("Blue: %s|%s|%s\n", el.Attr("data-name"), el.Text(), el.HTML())

	// Blue2
	el = h.Find("div", map[string]string{"class": " blue blue2 "})
	tt.Logf("Blue2: %s|%s|%s\n", el.Attr("data-name"), el.Text(), el.HTML())

}

package znet

import (
	"net/http"
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zutil"
)

func TestRenderProcessingSkipAfterAbort(t *testing.T) {
	tt := zlsgo.NewTest(t)
	c := &Context{
		stopHandle: zutil.NewBool(true),
		prevData: &PrevData{
			Code: zutil.NewInt32(http.StatusTeapot),
			Type: ContentTypePlain,
		},
		header: map[string][]string{},
	}

	c.String(http.StatusOK, "ok")

	tt.Equal(int32(http.StatusTeapot), c.prevData.Code.Load())
	tt.Equal(nil, c.render)
}

func TestRenderProcessingAllowAfterAbortWithoutCode(t *testing.T) {
	tt := zlsgo.NewTest(t)
	c := &Context{
		stopHandle: zutil.NewBool(true),
		prevData: &PrevData{
			Code: zutil.NewInt32(0),
			Type: ContentTypePlain,
		},
		header: map[string][]string{},
	}

	c.String(http.StatusOK, "ok")

	tt.Equal(int32(http.StatusOK), c.prevData.Code.Load())
	_, ok := c.render.(*renderString)
	tt.Equal(true, ok)
}

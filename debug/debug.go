package debug

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

func CopyReq(req *http.Request) *http.Request {
	r2 := &http.Request{}
	*r2 = *req

	if req.Body != nil {
		var b bytes.Buffer

		_, err := b.ReadFrom(req.Body)
		if err == nil {
			req.Body = ioutil.NopCloser(&b)
			r2.Body = ioutil.NopCloser(bytes.NewReader(b.Bytes()))
		}
	}

	return r2
}

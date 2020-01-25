package deployment

import (
	"github.com/nbio/st"
	"gopkg.in/h2non/gock.v1"
	"io/ioutil"
	"net/http"
	"testing"
)

// https://stackoverflow.com/questions/43240970/how-to-mock-http-client-do-method/43241303#43241303
// https://github.com/vektra/mockery
// https://github.com/dankinder/httpmock
// https://github.com/h2non/gock
// https://github.com/vektra/mockery
// github.com/stretchr/testify/mock
func TestHttp(test *testing.T) {
	test.Run("mock simple http", func(test *testing.T) {
		defer gock.Off()
		gock.New("http://foo.com").
			Get("/bar").
			Reply(200).
			JSON(map[string]string{"foo": "bar"})

		res, err := http.Get("http://foo.com/bar?query1=value1")
		st.Expect(test, err, nil)
		st.Expect(test, res.StatusCode, 200)

		body, _ := ioutil.ReadAll(res.Body)
		st.Expect(test, string(body)[:13], `{"foo":"bar"}`)
		// Verify that we don't have pending mocks
		st.Expect(test, gock.IsDone(), true)
	})
}

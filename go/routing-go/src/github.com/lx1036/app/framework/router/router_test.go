package router

import (
	"net/http"
	"reflect"
	"testing"
)
/*
https://github.com/julienschmidt/httprouter/blob/master/router_test.go
 */
func TestParams(test *testing.T) {
	params := Params{
		Param{"param1", "value1"},
		Param{"param2", "value2"},
		Param{"param3", "value3"},
	}

	for key, param := range params {
		if value := param.ByName(params[key].Key); value != params[key].Value {
			test.Errorf("Wrong value for %s got %s, want %s", params[key].Key, value, params[key].Value)
		}
	}

	if value := params.ByName("noKey"); value != "" {
		test.Errorf("Got %s, but want empty value", value)
	}
}

type mockResponseWriter struct {

}

func (mock *mockResponseWriter) Header() http.Header {
	return http.Header{}
}

func (mock *mockResponseWriter) Write(response []byte) (int, error)  {
	return len(response), nil
}

func (mock *mockResponseWriter) WriteHeader(statusCode int) {}

func TestRouter(test *testing.T) {
	router := New()
	routed := false
	router.Handle("GET", "/user/:name", func(writer http.ResponseWriter, request *http.Request, params Params) {
		routed = true
		want := Params{
			Param{Key: "name", Value: "lx1036"},
		}
		if !reflect.DeepEqual(want, params) {
			test.Fatalf("wrong values: want %v got %v", want, params)
		}
	})

	mockWriter := new(mockResponseWriter)
	request, _ := http.NewRequest("GET", "/user/lx1036", nil)
	router.ServeHTTP(mockWriter, request)

	if !routed {
		test.Fatal("routed failed")
	}
}

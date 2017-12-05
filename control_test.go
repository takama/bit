package bit

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type prm struct {
	Key, Value string
}

var params1 = []prm{
	{"name", "John"},
	{"age", "32"},
	{"gender", "M"},
}

var params2 = []prm{
	{"name", "Jane"},
	{"age", "33"},
	{"gender", "F"},
}

var testParamsData = `[{"Key":"name","Value":"John"},{"Key":"age","Value":"32"},{"Key":"gender","Value":"M"}]`
var testParamGzipData = []byte{
	31, 139, 8, 0, 0, 0, 0, 0, 0, 255, 138, 174, 86, 242, 78, 173, 84, 178, 82, 202,
	75, 204, 77, 85, 210, 81, 10, 75, 204, 41, 77, 85, 178, 82, 242, 202, 207, 200,
	83, 170, 213, 129, 201, 38, 166, 35, 75, 26, 27, 33, 73, 165, 167, 230, 165, 164,
	22, 33, 201, 250, 42, 213, 198, 2, 2, 0, 0, 255, 255, 196, 73, 247, 37, 87, 0, 0, 0,
}
var testStrData = "plain text"
var testStrGzipData = []byte{
	31, 139, 8, 0, 0, 0, 0, 0, 0, 255, 42, 200, 73, 204, 204, 83, 40,
	73, 173, 40, 1, 4, 0, 0, 255, 255, 184, 15, 11, 68, 10, 0, 0, 0,
}

func TestParamsQueryGet(t *testing.T) {

	p := make(Params, 0)
	c := &control{params: &p}
	for _, param := range params1 {
		c.Params().Set(param.Key, param.Value)
	}
	for _, param := range params1 {
		value, ok := c.Params().Get(param.Key)
		if !ok {
			t.Error("Expected ok, got false")
		}
		if value != param.Value {
			t.Error("Expected for", param.Key, ":", param.Value, ", got", value)
		}
	}
	for _, param := range params2 {
		c.Params().Set(param.Key, param.Value)
	}
	for _, param := range params2 {
		value := c.Query(param.Key)
		if value != param.Value {
			t.Error("Expected for", param.Key, ":", param.Value, ", got", value)
		}
	}
}

func TestWriterHeader(t *testing.T) {
	req, err := http.NewRequest("GET", "hello/:name", nil)
	if err != nil {
		t.Error(err)
	}
	trw := httptest.NewRecorder()
	c := NewControl(trw, req)
	request := c.Request()
	if request != req {
		t.Error("Expected", req.URL.String(), "got", request.URL)
	}
	trw.Header().Add("Test", "TestValue")
	c = NewControl(trw, req)
	expected := trw.Header().Get("Test")
	value := c.Header().Get("Test")
	if value != expected {
		t.Error("Expected", expected, "got", value)
	}
}

func TestWriterCode(t *testing.T) {
	c := new(control)
	// code transcends, must be less than 600
	c.Code(777)
	if c.code != 0 {
		t.Error("Expected code", "0", "got", c.code)
	}
	c.Code(404)
	if c.code != 404 {
		t.Error("Expected code", "404", "got", c.code)
	}
}

func TestGetCode(t *testing.T) {
	c := new(control)
	c.Code(http.StatusOK)
	code := c.GetCode()
	if code != http.StatusOK {
		t.Error("Expected code", http.StatusText(http.StatusOK), "got", code)
	}
}

func TestWrite(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Error(err)
	}

	// Write data using http.ResponseWriter
	trw := httptest.NewRecorder()
	c := NewControl(trw, req)
	c.WriteHeader(http.StatusAccepted)
	c.Write([]byte("Testing"))
	if trw.Code != http.StatusAccepted {
		t.Error("Expected", http.StatusAccepted, "got", trw.Code)
	}
	if trw.Body.String() != "Testing" {
		t.Error("Expected", "Testing", "got", trw.Body)
	}

	// Write plain text data
	trw = httptest.NewRecorder()
	c = NewControl(trw, req)
	c.Body("Hello")
	if trw.Body.String() != "Hello" {
		t.Error("Expected", "Hello", "got", trw.Body)
	}
	contentType := trw.Header().Get("Content-type")
	expected := "text/plain; charset=utf-8"
	if contentType != expected {
		t.Error("Expected", expected, "got", contentType)
	}

	// Write JSON compatible data
	trw = httptest.NewRecorder()
	c = NewControl(trw, req)
	c.Code(http.StatusOK)
	c.Body(params1)
	if trw.Body.String() != testParamsData {
		t.Error("Expected", testParamsData, "got", trw.Body)
	}
	contentType = trw.Header().Get("Content-type")
	expected = "application/json"
	if contentType != expected {
		t.Error("Expected", expected, "got", contentType)
	}

	// Write encoded string
	req.Header.Add("Accept-Encoding", "gzip, deflate")
	trw = httptest.NewRecorder()
	c = NewControl(trw, req)
	c.Code(http.StatusAccepted)
	c.Body(testStrData)
	if trw.Body.String() != string(testStrGzipData) {
		t.Error("Expected", testStrGzipData, "got", trw.Body)
	}
	contentEncoding := trw.Header().Get("Content-Encoding")
	expected = "gzip"
	if contentEncoding != expected {
		t.Error("Expected", expected, "got", contentEncoding)
	}

	// Write encoded struct data
	req.Header.Add("Accept-Encoding", "gzip, deflate")
	trw = httptest.NewRecorder()
	c = NewControl(trw, req)
	c.Code(http.StatusAccepted)
	c.Body(params1)
	if trw.Body.String() != string(testParamGzipData) {
		t.Error("Expected", testParamGzipData, "got", trw.Body)
	}
	contentEncoding = trw.Header().Get("Content-Encoding")
	expected = "gzip"
	if contentEncoding != expected {
		t.Error("Expected", expected, "got", contentEncoding)
	}

	// Try to write unexpected data type
	trw = httptest.NewRecorder()
	c = NewControl(trw, req)
	c.Body(func() {})
	if trw.Code != http.StatusInternalServerError {
		t.Error("Expected", http.StatusInternalServerError, "got", trw.Code)
	}
	expected = "application/json"
	if contentType != expected {
		t.Error("Expected", expected, "got", contentType)
	}
}

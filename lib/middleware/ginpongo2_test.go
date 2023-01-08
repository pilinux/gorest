package middleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pilinux/gorest/lib/middleware"
)

func TestStringFromContext(t *testing.T) {
	// create a new gin context
	c, _ := gin.CreateTestContext(nil)

	// set a value in the context
	c.Set("testKey", "testValue")

	// call the StringFromContext function with the gin context and the key to retrieve
	t.Logf("calling StringFromContext with string")
	result := middleware.StringFromContext(c, "testKey")

	// assert that the result is as expected
	if result != "testValue" {
		t.Errorf("expected 'testValue', got '%s'", result)
	}

	// call the StringFromContext function with the gin context and an empty string as the key
	t.Logf("calling StringFromContext with empty string")
	result = middleware.StringFromContext(c, "")

	// assert that the result is an empty string
	if result != "" {
		t.Errorf("expected '', got '%s'", result)
	}
}

func TestConvertContext(t *testing.T) {
	// create a test map
	testMap := map[string]interface{}{
		"key1": "value1",
		"key2": 2,
	}

	// call the ConvertContext function with the test map as input
	t.Logf("call the ConvertContext function with the test map as input")
	result := middleware.ConvertContext(testMap)

	// assert that the result is a pongo2.Context type and contains the expected values
	if result == nil {
		t.Errorf("expected a pongo2.Context, got nil")
	}
	if result["key1"] != "value1" {
		t.Errorf("expected 'value1', got '%s'", result["key1"])
	}
	if result["key2"] != 2 {
		t.Errorf("expected '2', got '%d'", result["key2"])
	}

	// call the ConvertContext function with nil as input
	t.Logf("call the ConvertContext function with nil as input")
	result = middleware.ConvertContext(nil)

	// assert that the result is nil
	if result != nil {
		t.Errorf("expected nil, got '%v'", result)
	}
}

func TestPongo2(t *testing.T) {
	testCases := []struct {
		name          string
		baseDirectory string
		template      string
		data          interface{}
		expectedCode  int
		expectedBody  string
	}{
		{
			"valid data",
			"templates",
			"index.html",
			map[string]interface{}{
				"message": "Hello, World!",
			},
			http.StatusOK,
			"Hello, World!",
		},

		{
			"missing template name",
			"templates",
			"",
			map[string]interface{}{
				"message": "Hello, World!",
			},
			http.StatusOK,
			"",
		},

		{
			"missing data",
			"templates",
			"index.html",
			nil,
			http.StatusOK,
			"",
		},

		{
			"invalid base directory",
			"invalid/path",
			"index.html",
			map[string]interface{}{
				"message": "Hello, World!",
			},
			http.StatusInternalServerError,
			"",
		},

		{
			"invalid template file",
			"templates",
			"invalid.html",
			map[string]interface{}{
				"message": "Hello, World!",
			},
			http.StatusInternalServerError,
			"",
		},

		{
			"panicking template execution",
			"templates",
			"index.html",
			map[string]interface{}{
				"message": "Hello, World!",
			},
			http.StatusOK,
			"",
		},
	}

	// create a new directory for testing
	err := os.Mkdir("templates", 0700)
	if err != nil {
		t.Error(err)
	}

	// download a file from a remote location and save it to the new directory
	fileUrl := strings.TrimSpace(os.Getenv("TEST_INDEX_HTML_URL"))
	err = downloadFile("templates/index.html", fileUrl)
	if err != nil {
		t.Error(err)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// set up a gin router and handler
			gin.SetMode(gin.TestMode)
			router := gin.New()
			err = router.SetTrustedProxies(nil)
			if err != nil {
				t.Errorf("failed to set trusted proxies to nil")
			}
			router.TrustedPlatform = "X-Real-Ip"

			router.Use(middleware.Pongo2(tc.baseDirectory))
			router.GET("/", func(c *gin.Context) {
				c.Set("template", tc.template)
				c.Set("data", tc.data)
				if tc.name == "panicking template execution" {
					panic("test panic")
				}
			})

			// create a request and response recorder
			w := httptest.NewRecorder()
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Errorf("failed to create an HTTP request")
				return
			}

			// call the handler
			router.ServeHTTP(w, req)

			// check the status code
			if w.Code != tc.expectedCode {
				t.Errorf("unexpected status code: got %v, want %v", w.Code, tc.expectedCode)
			}

			// check the response body
			if !strings.Contains(w.Body.String(), tc.expectedBody) {
				t.Errorf("unexpected response body: got %v, want %v", w.Body.String(), tc.expectedBody)
			}
		})
	}

	// remove the directory at the end of the test
	err = os.RemoveAll("templates")
	if err != nil {
		t.Error(err)
	}
}

// downloadFile will download a url and save it to a local file.
// It's efficient because it will write as it downloads and not
// load the whole file into memory.
func downloadFile(filepath string, url string) error {

	// create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

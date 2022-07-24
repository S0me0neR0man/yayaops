package server

import (
	"bytes"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlers(t *testing.T) {
	// определяем структуру теста
	type want struct {
		code        int
		body        string
		contentType string
	}
	// создаём массив тестов: имя и желаемый результат
	var tests = []struct {
		name        string
		url         string
		method      string
		contentType string
		body        string
		want        want
	}{
		{
			name:   "#1 positive",
			url:    "/update/counter/Alloc/124",
			method: http.MethodPost,
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name:   "#2 wrong method",
			url:    "/update/gauge/Alloc/124.6",
			method: http.MethodDelete,
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name:   "#3 invalid body",
			url:    "/update/gauge/BUGAlloc/none",
			method: http.MethodPost,
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name:   "#4 without id",
			url:    "/update/counter/",
			method: http.MethodPost,
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name:   "#5 without id",
			url:    "/update/gauge/",
			method: http.MethodPost,
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name:   "#6 wrong oper",
			url:    "/update/unknown/testCounter/100",
			method: http.MethodPost,
			want: want{
				code: http.StatusNotImplemented,
			},
		},
		{
			name:   "#7 set ",
			url:    "/update/gauge/val/100.3567",
			method: http.MethodPost,
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name:   "#8 get after set ",
			url:    "/value/gauge/val",
			method: http.MethodGet,
			want: want{
				code: http.StatusOK,
				body: "100.3567",
			},
		},
		// ------ INCREMENT 4
		{
			name:        "#9 JSON post ",
			url:         "/update/",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        "{\"id\":\"GCSys\",\"type\":\"counter\",\"delta\":1000}",
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name:        "#10 JSON post ",
			url:         "/update/",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        "{\"id\":\"GCSys\",\"type\":\"counter\",\"delta\":1000}",
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name:        "#11 JSON get ",
			url:         "/value/",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        "{\"id\":\"GCSys\",\"type\":\"counter\"}",
			want: want{
				code:        http.StatusOK,
				body:        "{\"id\":\"GCSys\",\"type\":\"counter\",\"delta\":2000}",
				contentType: "application/json",
			},
		},
		{
			name:        "#12 JSON get unknown counter",
			url:         "/value/",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        "{\"id\":\"NONE\",\"type\":\"counter\"}",
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name:        "#13 JSON get unknown ID",
			url:         "/value/",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        "{\"id\":\"NONE\",\"type\":\"gauge\"}",
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name:   "#14 JSON post unknown type",
			url:    "/update/",
			method: http.MethodPost,

			contentType: "application/json",
			body:        "{\"id\":\"GCSys\",\"type\":\"NONE\",\"delta\":2000}",
			want: want{
				code: http.StatusNotImplemented,
			},
		},
		{
			name:        "#15 JSON post ",
			url:         "/update/",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        "{\"id\":\"VAL\",\"type\":\"gauge\",\"value\":100.123}",
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
			},
		},
		{
			name:        "#16 JSON get ",
			url:         "/value/",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        "{\"id\":\"VAL\",\"type\":\"gauge\"}",
			want: want{
				code:        http.StatusOK,
				body:        "{\"id\":\"VAL\",\"type\":\"gauge\",\"value\":100.123}",
				contentType: "application/json",
			},
		},
		{
			name:        "#17 wrong JSON post ",
			url:         "/update/",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        "{\"id\":\"VALtoSET\",\"type\":\"gauge\"}",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "application/json",
			},
		},
		{
			name:        "#18 wrong JSON post ",
			url:         "/update/",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        "{\"id\":\"VALtoSET\",\"type\":\"counter\"}",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "application/json",
			},
		},
		{
			name:        "#19 wrong JSON post ",
			url:         "/update/",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        "{\"id\":\"TOG\",\"type\":\"gauge\",\"value\":1000.1}",
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
			},
		},
		{
			name:        "#20 wrong JSON post ",
			url:         "/update/",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        "{\"id\":\"TOG\",\"type\":\"counter\",\"delta\":1000}",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "application/json",
			},
		},
		{
			name:        "#21 wrong JSON post ",
			url:         "/update/",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        "{\"id\":\"TOG\",\"type\":\"counter\",\"value\":1000.1}",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "application/json",
			},
		},
		{
			name:        "#22 wrong JSON post ",
			url:         "/update/",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        "{\"id\":\"TOG\",\"type\":\"gauge\",\"delta\":1000}",
			want: want{
				code:        http.StatusBadRequest,
				contentType: "application/json",
			},
		},
		// {"id":"GCSys","type":"counter","delta":3807944}
	}
	s := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			request := httptest.NewRequest(tt.method, tt.url, bytes.NewBufferString(tt.body))
			if tt.contentType != "" {
				request.Header.Set("Content-Type", tt.contentType)
			}

			w := httptest.NewRecorder()

			router := mux.NewRouter()
			s.setHandlers(router)
			router.ServeHTTP(w, request)
			result := w.Result()

			assert.Equal(t, tt.want.code, result.StatusCode)
			if tt.want.contentType != "" {
				assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
			}
			value, err := ioutil.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			//var user User
			//err = json.Unmarshal(value, &user)
			//require.NoError(t, err)

			if tt.want.body != "" {
				assert.Equal(t, tt.want.body, string(value))
			}
		})
	}
}

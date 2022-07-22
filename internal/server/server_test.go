package server

import (
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
		response    string
		contentType string
	}
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name        string
		url         string
		method      string
		contentType string
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
				code: http.StatusNotAcceptable,
			},
		},
		{
			name:   "#3 invalid response",
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
				code:     http.StatusOK,
				response: "100.3567",
			},
		},
		// {"id":"GCSys","type":"counter","delta":3807944}
	}

	s := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.url, nil)
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

			if tt.want.response != "" {
				assert.Equal(t, tt.want.response, string(value))
			}
		})
	}
}

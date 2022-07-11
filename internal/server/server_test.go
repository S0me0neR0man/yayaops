package server

import (
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStatusHandler(t *testing.T) {
	// определяем структуру теста
	type want struct {
		url    string
		method string
		code   int
	}
	// создаём массив тестов: имя и желаемый результат
	tests := []struct {
		name string
		want want
	}{
		{
			name: "#1 positive",
			want: want{
				url:    "/update/counter/Alloc/124",
				method: http.MethodPost,
				code:   http.StatusOK,
			},
		},
		/*{
			name: "#2 wrong method",
			want: want{
				url:    "/update/gauge/Alloc/124.6",
				method: http.MethodDelete,
				code:   http.StatusNotAcceptable,
			},
		},*/
		{
			name: "#3 invalid value",
			want: want{
				url:    "/update/gauge/BUGAlloc/none",
				method: http.MethodPost,
				code:   http.StatusBadRequest,
			},
		},
		{
			name: "#4 without id",
			want: want{
				url:    "/update/counter/",
				method: http.MethodPost,
				code:   http.StatusNotFound,
			},
		},
		{
			name: "#5 without id",
			want: want{
				url:    "/update/gauge/",
				method: http.MethodPost,
				code:   http.StatusNotFound,
			},
		},
		{
			name: "#6 wrong oper",
			want: want{
				url:    "/update/unknown/testCounter/100",
				method: http.MethodPost,
				code:   http.StatusNotImplemented,
			},
		},
	}

	s := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.want.method, tt.want.url, nil)

			w := httptest.NewRecorder()

			router := mux.NewRouter()
			router.HandleFunc("/{oper}/{type}/{metric}/{value}", s.metricsPostHandler)
			router.ServeHTTP(w, request)

			//h := http.HandlerFunc(s.metricsPostHandler)
			//h.ServeHTTP(w, request)
			//res := w.Result()

			if w.Code != tt.want.code {
				t.Errorf("Expected status code %d, got %d", tt.want.code, w.Code)
			}

			// получаем и проверяем тело запроса
			/*
				defer res.Body.Close()
				resBody, err := io.ReadAll(res.Body)
				if err != nil {
					t.Fatal(err)
				}
				if string(resBody) != tt.want.response {
					t.Errorf("Expected body %s, got %s", tt.want.response, w.Body.String())
				}
			*/

			// заголовок ответа
			/*
				if res.Header.Get("Content-Type") != tt.want.contentType {
					t.Errorf("Expected Content-Type %s, got %s", tt.want.contentType, res.Header.Get("Content-Type"))
				}
			*/
		})
	}
}

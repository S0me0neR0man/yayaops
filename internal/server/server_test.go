package server

import (
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
				url:    "/update/gauge/Alloc/124.6",
				method: http.MethodPost,
				code:   200,
			},
		},
		{
			name: "#2 wrong method",
			want: want{
				url:    "/update/gauge/Alloc/124.6",
				method: http.MethodGet,
				code:   406,
			},
		},
		{
			name: "#3 invalid name",
			want: want{
				url:    "/update/gauge/BUGAlloc/124.6",
				method: http.MethodPost,
				code:   404,
			},
		},
		{
			name: "#4 invalid type",
			want: want{
				url:    "/update/Alloc/124.6",
				method: http.MethodPost,
				code:   404,
			},
		},
	}

	for _, tt := range tests {
		// запускаем каждый тест
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.want.method, tt.want.url, nil)

			// создаём новый Recorder
			w := httptest.NewRecorder()
			// определяем хендлер
			h := http.HandlerFunc(oneForAllHandler)
			// запускаем сервер
			h.ServeHTTP(w, request)
			res := w.Result()

			// проверяем код ответа
			if res.StatusCode != tt.want.code {
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

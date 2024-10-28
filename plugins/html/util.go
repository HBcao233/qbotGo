package html

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func response(w http.ResponseWriter, code int, message string, data any) error {
	w.Header().Add("content-type", "application/json")
	data, err := json.Marshal(map[string]any{
		"code":    code,
		"message": message,
		"data":    data,
	})
	if err != nil {
		return err
	}
	w.Write(data.([]byte))
	return nil
}
func success(w http.ResponseWriter, data any, message string) error {
	return response(w, 0, message, data)
}
func fail(w http.ResponseWriter, data any, message string) error {
	return response(w, 1, message, data)
}

func paramInt64(r *http.Request, name string) (int64, error) {
	s := r.URL.Query().Get(name)
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid semicolon separator in query")
	}
	return i, nil
}

func paramString(r *http.Request, name string) string {
	return r.URL.Query().Get(name)
}

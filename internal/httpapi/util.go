package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func jsonError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func parseID(r *http.Request) (int64, error) {
	idStr := strings.TrimSpace(r.URL.Query().Get("id"))
	return strconv.ParseInt(idStr, 10, 64) // вернёт ошибку только если id пуст или нет числа
}

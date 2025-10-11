package transport_test

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"port-service/internal/repository/inmem"
	"port-service/internal/services"
	"port-service/internal/transport"
	"github.com/stretchr/testify/require"
)

//go:embed testfixtures/ports_request.json
var portsRequest []byte

//go:embed testfixtures/ports_response.json
var portsResponse []byte

func TestUploadPorts(t *testing.T) {
	// Подсчёт ожидаемого количества уникальных портов из запроса
	requestPortsTotal, err := countJSONPorts(portsRequest)
	require.NoError(t, err)

	// Чистое состояние для каждого теста
	store := inmem.NewPortStore()
	portService := services.NewPortService(store)
	httpServer := transport.NewHttpServer(portService)

	req := httptest.NewRequest(http.MethodPost, "/ports", bytes.NewBuffer(portsRequest))
	w := httptest.NewRecorder()

	httpServer.UploadPorts(w, req)

	res := w.Result()
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, res.StatusCode)
	// Сравнение JSON по семантике, чтобы игнорировать перевод строки/пробелы/порядок ключей
	require.JSONEq(t, string(portsResponse), string(data))

	storedPortsTotal, err := portService.CountPorts(context.Background())
	require.NoError(t, err)
	require.Equal(t, requestPortsTotal, storedPortsTotal)
}

func TestUploadPorts_badJSON(t *testing.T) {
	// Чистое состояние для каждого теста
	store := inmem.NewPortStore()
	portService := services.NewPortService(store)
	httpServer := transport.NewHttpServer(portService)

	req := httptest.NewRequest(http.MethodPost, "/ports", bytes.NewBuffer([]byte("blabla")))
	w := httptest.NewRecorder()

	httpServer.UploadPorts(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusBadRequest, res.StatusCode)
}

// countJSONPorts возвращает количество ключей верхнего уровня в JSON-объекте
func countJSONPorts(portsJSON []byte) (int, error) {
	var ports map[string]struct{}
	if err := json.Unmarshal(portsJSON, &ports); err != nil {
		return 0, err
	}
	return len(ports), nil
}

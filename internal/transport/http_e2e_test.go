package transport_test

import (
	"bytes"
	"context"
	_"embed"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"port-service/internal/repository/inmem"
	"port-service/internal/services"
	"port-service/internal/transport"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

//go:embed testfixtures/ports_request.json
var portsRequest []byte

//go:embed testfixtures/ports_response.json
var portsResponse []byte

type HttpTestSuite struct {
	suite.Suite
	portService transport.PortService
	httpServer  transport.HttpServer
}

func NewHttpTestSuite() *HttpTestSuite {
	s := &HttpTestSuite{}
	store := inmem.NewPortStore()
	s.portService = services.NewPortService(store)
	s.httpServer = transport.NewHttpServer(s.portService)
	return s
}

// Каждый тест стартует с чистым состоянием стора
func (s *HttpTestSuite) SetupTest() {
	store := inmem.NewPortStore()
	s.portService = services.NewPortService(store)
	s.httpServer = transport.NewHttpServer(s.portService)
}

func TestHttpTestSuite(t *testing.T) {
	suite.Run(t, NewHttpTestSuite())
}

func (s *HttpTestSuite) TestUploadPorts() {
	// Подсчёт ожидаемого количества уникальных портов из запроса
	requestPortsTotal := countJSONPorts(s.T(), portsRequest)

	req := httptest.NewRequest(http.MethodPost, "/ports", bytes.NewBuffer(portsRequest))
	w := httptest.NewRecorder()

	s.httpServer.UploadPorts(w, req)

	res := w.Result()
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	require.NoError(s.T(), err)

	require.Equal(s.T(), http.StatusOK, res.StatusCode)
	// Сравниваем JSON по семантике, чтобы игнорировать перевод строки/пробелы/порядок ключей
	require.JSONEq(s.T(), string(portsResponse), string(data))

	storedPortsTotal, err := s.portService.CountPorts(context.Background())
	require.NoError(s.T(), err)
	require.Equal(s.T(), requestPortsTotal, storedPortsTotal)
}

func (s *HttpTestSuite) TestUploadPorts_badJSON() {
	req := httptest.NewRequest(http.MethodPost, "/ports", bytes.NewBuffer([]byte("blabla")))
	w := httptest.NewRecorder()

	s.httpServer.UploadPorts(w, req)

	res := w.Result()
	defer res.Body.Close()

	require.Equal(s.T(), http.StatusBadRequest, res.StatusCode)
}

func countJSONPorts(t *testing.T, portsJSON []byte) int {
	t.Helper()
	var ports map[string]struct{}
	err := json.Unmarshal(portsJSON, &ports)
	require.NoError(t, err)
	return len(ports)
}

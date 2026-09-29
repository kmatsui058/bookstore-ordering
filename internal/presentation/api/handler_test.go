package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kmatsui058/bookstore-ordering/internal/presentation/api"
	"github.com/kmatsui058/bookstore-ordering/internal/presentation/api/gen"
)

const validPlaceOrder = `{"customerID":"customer-1","lines":[{"bookID":"book-1","quantity":2,"unitPrice":1500}],"paymentMethod":"CREDIT_CARD"}`

// client は、テスト用に起動した注文 API に JSON のリクエストを送る。
type client struct {
	t   *testing.T
	srv *httptest.Server
}

func newClient(t *testing.T) client {
	t.Helper()
	srv := httptest.NewServer(api.NewHTTPHandler(slog.New(slog.NewTextHandler(io.Discard, nil))))
	t.Cleanup(srv.Close)
	return client{t: t, srv: srv}
}

// do は、リクエストを送り、ステータスコードと JSON の本文を返す。
func (c client) do(method, path, body string) (int, map[string]any) {
	c.t.Helper()
	req, err := http.NewRequest(method, c.srv.URL+path, bytes.NewBufferString(body))
	require.NoError(c.t, err)
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	require.NoError(c.t, err)
	defer res.Body.Close()
	var got map[string]any
	require.NoError(c.t, json.NewDecoder(res.Body).Decode(&got))
	return res.StatusCode, got
}

// placeOrder は、受付済みの注文を1件作り、その注文ID を返す。
func (c client) placeOrder() string {
	c.t.Helper()
	status, body := c.do(http.MethodPost, "/orders", validPlaceOrder)
	require.Equal(c.t, http.StatusCreated, status)
	return body["orderID"].(string)
}

func TestHandler_注文API(t *testing.T) {
	tests := []struct {
		name       string
		given      []string
		method     string
		path       string
		body       string
		wantStatus int
		wantField  string
		wantValue  any
	}{
		{
			name:   "Given: 何もない, When: 正しい内容で POST /orders, Then: 201 で受付済みの注文が返る",
			method: http.MethodPost, path: "/orders", body: validPlaceOrder,
			wantStatus: http.StatusCreated, wantField: "status", wantValue: string(gen.OrderStatusPLACED),
		},
		{
			name:   "Given: 何もない, When: 注文明細が0件で POST /orders, Then: 400 になる",
			method: http.MethodPost, path: "/orders", body: `{"customerID":"customer-1","lines":[],"paymentMethod":"CREDIT_CARD"}`,
			wantStatus: http.StatusBadRequest, wantField: "code", wantValue: "INVALID_ARGUMENT",
		},
		{
			name:   "Given: 何もない, When: 数量0の注文明細で POST /orders, Then: 400 になる",
			method: http.MethodPost, path: "/orders", body: `{"customerID":"customer-1","lines":[{"bookID":"book-1","quantity":0,"unitPrice":100}],"paymentMethod":"CREDIT_CARD"}`,
			wantStatus: http.StatusBadRequest, wantField: "code", wantValue: "INVALID_ARGUMENT",
		},
		{
			name:   "Given: 何もない, When: 定義にない支払方法で POST /orders, Then: 400 になる",
			method: http.MethodPost, path: "/orders", body: `{"customerID":"customer-1","lines":[{"bookID":"book-1","quantity":1,"unitPrice":100}],"paymentMethod":"POINT"}`,
			wantStatus: http.StatusBadRequest, wantField: "code", wantValue: "INVALID_ARGUMENT",
		},
		{
			name:   "Given: 受付済みの注文, When: GET /orders/{orderID}, Then: 200 で受付済みの注文が返る",
			method: http.MethodGet, path: "/orders/{orderID}",
			wantStatus: http.StatusOK, wantField: "status", wantValue: string(gen.OrderStatusPLACED),
		},
		{
			name:   "Given: 受付済みの注文, When: キャンセルする, Then: 200 でキャンセルの注文が返る",
			method: http.MethodPost, path: "/orders/{orderID}/cancel",
			wantStatus: http.StatusOK, wantField: "status", wantValue: string(gen.OrderStatusCANCELLED),
		},
		{
			name:  "Given: 確定の注文, When: キャンセルする, Then: 200 でキャンセルの注文が返る",
			given: []string{"/confirm"}, method: http.MethodPost, path: "/orders/{orderID}/cancel",
			wantStatus: http.StatusOK, wantField: "status", wantValue: string(gen.OrderStatusCANCELLED),
		},
		{
			name:  "Given: 出荷済みの注文, When: キャンセルする, Then: 409 になる",
			given: []string{"/confirm", "/ship"}, method: http.MethodPost, path: "/orders/{orderID}/cancel",
			wantStatus: http.StatusConflict, wantField: "code", wantValue: "INVALID_STATE",
		},
		{
			name:  "Given: キャンセルの注文, When: もう一度キャンセルする, Then: 409 になる",
			given: []string{"/cancel"}, method: http.MethodPost, path: "/orders/{orderID}/cancel",
			wantStatus: http.StatusConflict, wantField: "code", wantValue: "INVALID_STATE",
		},
		{
			name:   "Given: 受付済みの注文, When: 確定する, Then: 200 で確定の注文が返る",
			method: http.MethodPost, path: "/orders/{orderID}/confirm",
			wantStatus: http.StatusOK, wantField: "status", wantValue: string(gen.OrderStatusCONFIRMED),
		},
		{
			name:   "Given: 受付済みの注文, When: 出荷済みにする, Then: 409 になる",
			method: http.MethodPost, path: "/orders/{orderID}/ship",
			wantStatus: http.StatusConflict, wantField: "code", wantValue: "INVALID_STATE",
		},
		{
			name:  "Given: 確定の注文, When: 出荷済みにする, Then: 200 で出荷済みの注文が返る",
			given: []string{"/confirm"}, method: http.MethodPost, path: "/orders/{orderID}/ship",
			wantStatus: http.StatusOK, wantField: "status", wantValue: string(gen.OrderStatusSHIPPED),
		},
		{
			name:   "Given: ない注文ID, When: キャンセルする, Then: 404 になる",
			method: http.MethodPost, path: "/orders/00000000-0000-0000-0000-000000000000/cancel",
			wantStatus: http.StatusNotFound, wantField: "code", wantValue: "NOT_FOUND",
		},
		{
			name:   "Given: ない注文ID, When: GET /orders/{orderID}, Then: 404 になる",
			method: http.MethodGet, path: "/orders/00000000-0000-0000-0000-000000000000",
			wantStatus: http.StatusNotFound, wantField: "code", wantValue: "NOT_FOUND",
		},
		{
			name:   "Given: UUID ではない注文ID, When: GET /orders/{orderID}, Then: 400 になる",
			method: http.MethodGet, path: "/orders/not-a-uuid",
			wantStatus: http.StatusBadRequest, wantField: "code", wantValue: "INVALID_ARGUMENT",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newClient(t)
			orderID := c.placeOrder()
			for _, g := range tt.given {
				status, _ := c.do(http.MethodPost, "/orders/"+orderID+g, "")
				require.Equal(t, http.StatusOK, status)
			}
			path := bytes.ReplaceAll([]byte(tt.path), []byte("{orderID}"), []byte(orderID))

			status, body := c.do(tt.method, string(path), tt.body)

			assert.Equal(t, tt.wantStatus, status)
			assert.Equal(t, tt.wantValue, body[tt.wantField])
		})
	}
}

// Package api は、注文 API（契約リポジトリの OpenAPI の定義）のリクエストをユースケースにつなぐプレゼンテーション層のパッケージ。
// リクエストをドメインの型に変換してユースケースを呼び、結果とエラーを API のレスポンスに変換する。業務のルールは持たない。
package api

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/app/command"
	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/app/query"
	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/domain/order"
	"github.com/kmatsui058/bookstore-ordering/internal/presentation/api/gen"
	"github.com/kmatsui058/bookstore-ordering/internal/shared/domain/event"
)

// OrderCommands は、ハンドラーが使う注文のコマンドのユースケース。
type OrderCommands interface {
	PlaceOrder(ctx context.Context, in command.PlaceOrderInput) (*order.Aggregate, error)
	ConfirmOrder(ctx context.Context, orderID order.OrderID, correlationID event.CorrelationID) (*order.Aggregate, error)
	CancelOrder(ctx context.Context, orderID order.OrderID, correlationID event.CorrelationID) (*order.Aggregate, error)
	MarkOrderShipped(ctx context.Context, orderID order.OrderID, correlationID event.CorrelationID) (*order.Aggregate, error)
}

// OrderQueries は、ハンドラーが使う注文の読み取りのユースケース。
type OrderQueries interface {
	GetOrder(ctx context.Context, orderID order.OrderID) (query.Order, error)
}

// Handler は、生成したサーバーのインターフェース（gen.StrictServerInterface）を実装する注文 API のハンドラー。
type Handler struct {
	commands OrderCommands
	queries  OrderQueries
}

var _ gen.StrictServerInterface = (*Handler)(nil)

// NewHandler は、注文 API のハンドラーを作る。
func NewHandler(commands OrderCommands, queries OrderQueries) *Handler {
	return &Handler{commands: commands, queries: queries}
}

// PlaceOrder は、POST /orders（注文する）を処理する。
func (h *Handler) PlaceOrder(ctx context.Context, request gen.PlaceOrderRequestObject) (gen.PlaceOrderResponseObject, error) {
	in, err := toPlaceOrderInput(request.Body, newCorrelationID())
	if err != nil {
		return gen.PlaceOrder400JSONResponse{BadRequestJSONResponse: gen.BadRequestJSONResponse(invalidArgument(err))}, nil
	}
	agg, err := h.commands.PlaceOrder(ctx, in)
	if errors.Is(err, order.ErrInvalidArgument) {
		return gen.PlaceOrder400JSONResponse{BadRequestJSONResponse: gen.BadRequestJSONResponse(invalidArgument(err))}, nil
	}
	if err != nil {
		return nil, err
	}
	return gen.PlaceOrder201JSONResponse(toAPIOrder(agg)), nil
}

// GetOrder は、GET /orders/{orderID}（注文を取得する）を処理する。
func (h *Handler) GetOrder(ctx context.Context, request gen.GetOrderRequestObject) (gen.GetOrderResponseObject, error) {
	o, err := h.queries.GetOrder(ctx, order.OrderID(request.OrderID))
	if errors.Is(err, order.ErrNotFound) {
		return gen.GetOrder404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse(notFound(err))}, nil
	}
	if err != nil {
		return nil, err
	}
	return gen.GetOrder200JSONResponse(toAPIOrderFromView(o)), nil
}

// ConfirmOrder は、POST /orders/{orderID}/confirm（確定する）を処理する。
func (h *Handler) ConfirmOrder(ctx context.Context, request gen.ConfirmOrderRequestObject) (gen.ConfirmOrderResponseObject, error) {
	agg, err := h.commands.ConfirmOrder(ctx, order.OrderID(request.OrderID), newCorrelationID())
	switch {
	case err == nil:
		return gen.ConfirmOrder200JSONResponse(toAPIOrder(agg)), nil
	case errors.Is(err, order.ErrNotFound):
		return gen.ConfirmOrder404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse(notFound(err))}, nil
	case isConflict(err):
		return gen.ConfirmOrder409JSONResponse(conflict(err)), nil
	default:
		return nil, err
	}
}

// CancelOrder は、POST /orders/{orderID}/cancel（キャンセルする）を処理する。
func (h *Handler) CancelOrder(ctx context.Context, request gen.CancelOrderRequestObject) (gen.CancelOrderResponseObject, error) {
	agg, err := h.commands.CancelOrder(ctx, order.OrderID(request.OrderID), newCorrelationID())
	switch {
	case err == nil:
		return gen.CancelOrder200JSONResponse(toAPIOrder(agg)), nil
	case errors.Is(err, order.ErrNotFound):
		return gen.CancelOrder404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse(notFound(err))}, nil
	case isConflict(err):
		return gen.CancelOrder409JSONResponse(conflict(err)), nil
	default:
		return nil, err
	}
}

// MarkOrderShipped は、POST /orders/{orderID}/ship（出荷済みにする）を処理する。
func (h *Handler) MarkOrderShipped(ctx context.Context, request gen.MarkOrderShippedRequestObject) (gen.MarkOrderShippedResponseObject, error) {
	agg, err := h.commands.MarkOrderShipped(ctx, order.OrderID(request.OrderID), newCorrelationID())
	switch {
	case err == nil:
		return gen.MarkOrderShipped200JSONResponse(toAPIOrder(agg)), nil
	case errors.Is(err, order.ErrNotFound):
		return gen.MarkOrderShipped404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse(notFound(err))}, nil
	case isConflict(err):
		return gen.MarkOrderShipped409JSONResponse(conflict(err)), nil
	default:
		return nil, err
	}
}

// newCorrelationID は、API のリクエストを起点とする業務の流れの相関 ID を採番する。
func newCorrelationID() event.CorrelationID {
	return event.CorrelationID(uuid.NewString())
}

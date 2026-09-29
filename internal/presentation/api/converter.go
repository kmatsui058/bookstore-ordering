package api

import (
	"errors"
	"fmt"

	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/app/command"
	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/app/query"
	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/domain/order"
	"github.com/kmatsui058/bookstore-ordering/internal/presentation/api/gen"
	"github.com/kmatsui058/bookstore-ordering/internal/shared/domain/event"
	"github.com/kmatsui058/bookstore-ordering/internal/shared/domain/id"
)

// errMissingBody は、リクエストの本文がないことを表す。
var errMissingBody = errors.New("リクエストの本文がありません")

// toPlaceOrderInput は、注文するリクエストをユースケースの入力に変換する。値のルールはドメインの型を作るときに確かめる。
func toPlaceOrderInput(body *gen.PlaceOrderJSONRequestBody, correlationID event.CorrelationID) (command.PlaceOrderInput, error) {
	if body == nil {
		return command.PlaceOrderInput{}, errMissingBody
	}
	lines := make([]order.OrderLine, 0, len(body.Lines))
	for i, l := range body.Lines {
		line, err := order.NewOrderLine(id.BookID(l.BookID), l.Quantity, l.UnitPrice)
		if err != nil {
			return command.PlaceOrderInput{}, fmt.Errorf("注文明細 %d 件目: %w", i+1, err)
		}
		lines = append(lines, line)
	}
	paymentMethod, err := toDomainPaymentMethod(body.PaymentMethod)
	if err != nil {
		return command.PlaceOrderInput{}, err
	}
	return command.PlaceOrderInput{
		CustomerID:    id.CustomerID(body.CustomerID),
		Lines:         lines,
		PaymentMethod: paymentMethod,
		CorrelationID: correlationID,
	}, nil
}

// toAPIOrder は、集約を API の注文に変換する。
func toAPIOrder(agg *order.Aggregate) gen.Order {
	return toAPIOrderFromView(query.NewOrder(agg))
}

// toAPIOrderFromView は、読み取り用のモデルを API の注文に変換する。
func toAPIOrderFromView(o query.Order) gen.Order {
	lines := make([]gen.OrderLine, 0, len(o.Lines))
	for _, l := range o.Lines {
		lines = append(lines, gen.OrderLine{BookID: l.BookID.String(), Quantity: l.Quantity, UnitPrice: l.UnitPrice})
	}
	return gen.Order{
		OrderID:       o.OrderID.UUID(),
		CustomerID:    o.CustomerID.String(),
		Lines:         lines,
		PaymentMethod: toAPIPaymentMethod(o.PaymentMethod),
		Status:        toAPIOrderStatus(o.Status),
	}
}

// toAPIOrderStatus は、注文ステータスを API の値に変換する。
func toAPIOrderStatus(s order.OrderStatus) gen.OrderStatus {
	switch s {
	case order.OrderStatusPlaced:
		return gen.OrderStatusPLACED
	case order.OrderStatusConfirmed:
		return gen.OrderStatusCONFIRMED
	case order.OrderStatusShipped:
		return gen.OrderStatusSHIPPED
	case order.OrderStatusCancelled:
		return gen.OrderStatusCANCELLED
	}
	return gen.OrderStatus(s)
}

// toAPIPaymentMethod は、支払方法を API の値に変換する。
func toAPIPaymentMethod(p order.PaymentMethod) gen.PaymentMethod {
	switch p {
	case order.PaymentMethodCreditCard:
		return gen.PaymentMethodCREDITCARD
	case order.PaymentMethodBankTransfer:
		return gen.PaymentMethodBANKTRANSFER
	case order.PaymentMethodCashOnDelivery:
		return gen.PaymentMethodCASHONDELIVERY
	}
	return gen.PaymentMethod(p)
}

// toDomainPaymentMethod は、API の支払方法をドメインの値に変換する。定義にない値はエラーにする。
func toDomainPaymentMethod(p gen.PaymentMethod) (order.PaymentMethod, error) {
	switch p {
	case gen.PaymentMethodCREDITCARD:
		return order.PaymentMethodCreditCard, nil
	case gen.PaymentMethodBANKTRANSFER:
		return order.PaymentMethodBankTransfer, nil
	case gen.PaymentMethodCASHONDELIVERY:
		return order.PaymentMethodCashOnDelivery, nil
	}
	return "", fmt.Errorf("%w: %q", order.ErrInvalidPaymentMethod, p)
}

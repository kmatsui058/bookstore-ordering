package order

import "github.com/kmatsui058/bookstore-ordering/internal/shared/domain/id"

// OrderConfirmed は、注文が確定したことを表すドメインイベント（注文確定。LikeC4 ビュー bkst_ordr_ordm_ord_cfmd）。
// 請求コンテキストは、このイベントを受けて請求を発行する。
type OrderConfirmed struct {
	orderEvent
	customerID    id.CustomerID
	paymentMethod PaymentMethod
	totalAmount   int
}

var _ Event = OrderConfirmed{}

// EventType は、注文確定の種類を返す。
func (e OrderConfirmed) EventType() EventType {
	return EventTypeOrderConfirmed
}

// CustomerID は、注文した顧客の顧客ID を返す。
func (e OrderConfirmed) CustomerID() id.CustomerID {
	return e.customerID
}

// PaymentMethod は、支払方法を返す。
func (e OrderConfirmed) PaymentMethod() PaymentMethod {
	return e.paymentMethod
}

// TotalAmount は、注文明細の単価 × 数量の合計（円）を返す。
func (e OrderConfirmed) TotalAmount() int {
	return e.totalAmount
}

// apply は、注文を確定にする。
func (e OrderConfirmed) apply(a *Aggregate) {
	a.status = OrderStatusConfirmed
}

package order

import "github.com/kmatsui058/bookstore-ordering/internal/shared/domain/id"

// OrderPlaced は、注文が受け付けられたことを表すドメインイベント（注文済み。LikeC4 ビュー bkst_ordr_ordm_ord_plcd）。
// 注文の出発点なので、イベントの列から注文を復元するのに要る注文の中身（顧客ID・注文明細一覧・支払方法）も持つ。
type OrderPlaced struct {
	orderEvent
	customerID    id.CustomerID
	lines         []OrderLine
	paymentMethod PaymentMethod
}

var _ Event = OrderPlaced{}

// EventType は、注文済みの種類を返す。
func (e OrderPlaced) EventType() EventType {
	return EventTypeOrderPlaced
}

// CustomerID は、注文した顧客の顧客ID を返す。
func (e OrderPlaced) CustomerID() id.CustomerID {
	return e.customerID
}

// Lines は、注文明細一覧の複製を返す。
func (e OrderPlaced) Lines() []OrderLine {
	return append([]OrderLine(nil), e.lines...)
}

// PaymentMethod は、支払方法を返す。
func (e OrderPlaced) PaymentMethod() PaymentMethod {
	return e.paymentMethod
}

// apply は、注文を受付済みにし、注文の中身を当てはめる。
func (e OrderPlaced) apply(a *Aggregate) {
	a.orderID = e.orderID
	a.customerID = e.customerID
	a.lines = e.Lines()
	a.paymentMethod = e.paymentMethod
	a.status = OrderStatusPlaced
}

package order

// OrderCancelled は、注文がキャンセルされたことを表すドメインイベント（注文キャンセル。LikeC4 ビュー bkst_ordr_ordm_ord_cncd）。
type OrderCancelled struct {
	orderEvent
}

var _ Event = OrderCancelled{}

// EventType は、注文キャンセルの種類を返す。
func (e OrderCancelled) EventType() EventType {
	return EventTypeOrderCancelled
}

// apply は、注文をキャンセルにする。
func (e OrderCancelled) apply(a *Aggregate) {
	a.status = OrderStatusCancelled
}

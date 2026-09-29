package order

// OrderShipped は、注文の書籍が配送業者に渡されたことを表すドメインイベント（注文出荷。LikeC4 ビュー bkst_ordr_ordm_ord_shpd）。
type OrderShipped struct {
	orderEvent
}

var _ Event = OrderShipped{}

// EventType は、注文出荷の種類を返す。
func (e OrderShipped) EventType() EventType {
	return EventTypeOrderShipped
}

// apply は、注文を出荷済みにする。
func (e OrderShipped) apply(a *Aggregate) {
	a.status = OrderStatusShipped
}

package order

// OrderStatus は、注文がいまどの状態にあるか（注文ステータス。LikeC4 ビュー bkst_ordr_ordm_ost）を表す列挙型。
// 値は契約リポジトリのモデルの列挙値と同じ文字列にする。
type OrderStatus string

const (
	// OrderStatusPlaced は、受付済み。
	OrderStatusPlaced OrderStatus = "PLACED"
	// OrderStatusConfirmed は、確定。
	OrderStatusConfirmed OrderStatus = "CONFIRMED"
	// OrderStatusShipped は、出荷済み。
	OrderStatusShipped OrderStatus = "SHIPPED"
	// OrderStatusCancelled は、キャンセル。
	OrderStatusCancelled OrderStatus = "CANCELLED"
)

// AllOrderStatuses は、注文ステータスのすべての値を返す。値を足したらここにも足す。
func AllOrderStatuses() []OrderStatus {
	return []OrderStatus{OrderStatusPlaced, OrderStatusConfirmed, OrderStatusShipped, OrderStatusCancelled}
}

// String は、注文ステータスの値を返す。
func (s OrderStatus) String() string {
	return string(s)
}

// isBeforeShipment は、出荷前（受付済み・確定）の状態かどうかを返す。
func (s OrderStatus) isBeforeShipment() bool {
	switch s {
	case OrderStatusPlaced, OrderStatusConfirmed:
		return true
	case OrderStatusShipped, OrderStatusCancelled:
		return false
	}
	return false
}

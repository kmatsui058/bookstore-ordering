package order

import "github.com/google/uuid"

// OrderID は、注文を特定する識別子（注文ID。LikeC4 ビュー bkst_ordr_ordm_oid）。注文を受け付けたときに決まる。
type OrderID uuid.UUID

// NewOrderID は、新しい注文ID を採番する。
func NewOrderID() OrderID {
	return OrderID(uuid.New())
}

// String は、注文ID を UUID の文字列表現で返す。
func (o OrderID) String() string {
	return uuid.UUID(o).String()
}

// UUID は、注文ID を UUID として返す。
func (o OrderID) UUID() uuid.UUID {
	return uuid.UUID(o)
}

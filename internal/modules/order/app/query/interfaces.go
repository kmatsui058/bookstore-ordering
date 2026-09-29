package query

import (
	"context"

	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/domain/order"
)

// SnapshotReader は、コマンドの側が保存した注文のスナップショットを読む。アプリケーション層が要求し、インフラ層が実装する。
type SnapshotReader interface {
	// FindOrder は、注文のスナップショットを読む。ないときは order.ErrNotFound を返す。
	FindOrder(ctx context.Context, orderID order.OrderID) (Order, error)
}

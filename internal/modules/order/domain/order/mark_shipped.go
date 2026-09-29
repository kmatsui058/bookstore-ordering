package order

import (
	"time"

	"github.com/kmatsui058/bookstore-ordering/internal/shared/domain/event"
)

// MarkShipped は、コマンド「出荷済みにする」（MarkOrderShipped。LikeC4 ビュー bkst_ordr_ordm_ord_shp）を実行する。
// 出荷済みにできるのは確定した注文だけで、成功すると注文出荷（OrderShipped）を発行する。
func (a *Aggregate) MarkShipped(correlationID event.CorrelationID, occurredAt time.Time) error {
	if a.status != OrderStatusConfirmed {
		return ErrNotShippable
	}
	a.raise(OrderShipped{
		orderEvent: orderEvent{orderID: a.orderID, metadata: a.nextMetadata(correlationID, occurredAt)},
	})
	return nil
}

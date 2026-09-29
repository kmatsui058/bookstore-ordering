package order

import (
	"time"

	"github.com/kmatsui058/bookstore-ordering/internal/shared/domain/event"
)

// Cancel は、コマンド「キャンセルする」（CancelOrder。LikeC4 ビュー bkst_ordr_ordm_ord_cnc）を実行する。
// キャンセルできるのは出荷前（受付済み・確定）の注文だけで、出荷済みやキャンセル済みの注文には使えない。
// 成功すると注文キャンセル（OrderCancelled）を発行する。
func (a *Aggregate) Cancel(correlationID event.CorrelationID, occurredAt time.Time) error {
	if !a.status.isBeforeShipment() {
		return ErrNotCancellable
	}
	return a.raise(OrderCancelled{
		orderEvent: orderEvent{orderID: a.orderID, metadata: a.nextMetadata(correlationID, occurredAt)},
	})
}

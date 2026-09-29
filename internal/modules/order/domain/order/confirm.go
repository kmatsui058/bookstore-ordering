package order

import (
	"time"

	"github.com/kmatsui058/bookstore-ordering/internal/shared/domain/event"
)

// Confirm は、コマンド「確定する」（ConfirmOrder。LikeC4 ビュー bkst_ordr_ordm_ord_cfm）を実行する。
// 確定できるのは受付済みの注文だけで、成功すると注文確定（OrderConfirmed）を合計金額とともに発行する。
func (a *Aggregate) Confirm(correlationID event.CorrelationID, occurredAt time.Time) error {
	if a.status != OrderStatusPlaced {
		return ErrNotConfirmable
	}
	a.raise(OrderConfirmed{
		orderEvent:    orderEvent{orderID: a.orderID, metadata: a.nextMetadata(correlationID, occurredAt)},
		customerID:    a.customerID,
		paymentMethod: a.paymentMethod,
		totalAmount:   a.TotalAmount(),
	})
	return nil
}

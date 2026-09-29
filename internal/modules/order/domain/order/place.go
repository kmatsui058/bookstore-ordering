package order

import (
	"time"

	"github.com/kmatsui058/bookstore-ordering/internal/shared/domain/event"
	"github.com/kmatsui058/bookstore-ordering/internal/shared/domain/id"
)

// Place は、コマンド「注文する」（PlaceOrder。LikeC4 ビュー bkst_ordr_ordm_ord_plc）を実行し、受付済みの注文を作る。
// 顧客ID は必須、注文明細は1件以上、支払方法は定義された値でなければならない。成功すると注文済み（OrderPlaced）を発行する。
func Place(
	orderID OrderID,
	customerID id.CustomerID,
	lines []OrderLine,
	paymentMethod PaymentMethod,
	correlationID event.CorrelationID,
	occurredAt time.Time,
	bus *Bus,
) (*Aggregate, error) {
	if customerID == "" {
		return nil, ErrMissingCustomerID
	}
	if len(lines) == 0 {
		return nil, ErrNoLines
	}
	for _, l := range lines {
		if err := l.validate(); err != nil {
			return nil, err
		}
	}
	if !paymentMethod.isValid() {
		return nil, ErrInvalidPaymentMethod
	}
	a := &Aggregate{bus: bus}
	placed := OrderPlaced{
		orderEvent:    orderEvent{orderID: orderID, metadata: a.nextMetadata(correlationID, occurredAt)},
		customerID:    customerID,
		lines:         append([]OrderLine(nil), lines...),
		paymentMethod: paymentMethod,
	}
	a.raise(placed)
	return a, nil
}

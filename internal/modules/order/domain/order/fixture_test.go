package order_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/domain/order"
	"github.com/kmatsui058/bookstore-ordering/internal/shared/domain/event"
	"github.com/kmatsui058/bookstore-ordering/internal/shared/domain/id"
)

var (
	testCorrelationID = event.CorrelationID("corr-1")
	testNow           = time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
)

// newLine は、テスト用の注文明細を作る。
func newLine(t *testing.T, bookID string, quantity, unitPrice int) order.OrderLine {
	t.Helper()
	l, err := order.NewOrderLine(id.BookID(bookID), quantity, unitPrice)
	require.NoError(t, err)
	return l
}

// placedOrder は、コマンド「注文する」で受付済みの注文を作る。
func placedOrder(t *testing.T) *order.Aggregate {
	t.Helper()
	a, err := order.Place(
		order.NewOrderID(),
		id.CustomerID("customer-1"),
		[]order.OrderLine{newLine(t, "book-1", 2, 1200), newLine(t, "book-2", 1, 800)},
		order.PaymentMethodCreditCard,
		testCorrelationID,
		testNow,
		order.NewBus(),
	)
	require.NoError(t, err)
	return a
}

// confirmedOrder は、受付済みの注文を確定して確定の注文を作る。
func confirmedOrder(t *testing.T) *order.Aggregate {
	t.Helper()
	a := placedOrder(t)
	require.NoError(t, a.Confirm(testCorrelationID, testNow))
	return a
}

// shippedOrder は、確定の注文を出荷済みにして出荷済みの注文を作る。
func shippedOrder(t *testing.T) *order.Aggregate {
	t.Helper()
	a := confirmedOrder(t)
	require.NoError(t, a.MarkShipped(testCorrelationID, testNow))
	return a
}

// cancelledOrder は、受付済みの注文をキャンセルしてキャンセルの注文を作る。
func cancelledOrder(t *testing.T) *order.Aggregate {
	t.Helper()
	a := placedOrder(t)
	require.NoError(t, a.Cancel(testCorrelationID, testNow))
	return a
}

// eventTypes は、イベントの種類の列を返す。
func eventTypes(events []order.Event) []order.EventType {
	types := make([]order.EventType, 0, len(events))
	for _, e := range events {
		types = append(types, e.EventType())
	}
	return types
}

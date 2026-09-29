package order_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/domain/order"
	"github.com/kmatsui058/bookstore-ordering/internal/shared/domain/event"
)

// command は、注文に対するコマンドの実行を表す。
type command func(a *order.Aggregate) error

var (
	confirm     command = func(a *order.Aggregate) error { return a.Confirm(testCorrelationID, testNow) }
	cancel      command = func(a *order.Aggregate) error { return a.Cancel(testCorrelationID, testNow) }
	markShipped command = func(a *order.Aggregate) error { return a.MarkShipped(testCorrelationID, testNow) }
)

func TestAggregate_StateTransitions(t *testing.T) {
	tests := []struct {
		name          string
		given         func(t *testing.T) *order.Aggregate
		when          command
		wantErr       error
		wantStatus    order.OrderStatus
		wantEventType order.EventType
	}{
		{
			name:  "Given: 受付済みの注文, When: 確定する, Then: 確定になり注文確定が発行される",
			given: placedOrder, when: confirm,
			wantStatus: order.OrderStatusConfirmed, wantEventType: order.EventTypeOrderConfirmed,
		},
		{
			name:  "Given: 確定の注文, When: 確定する, Then: 確定できるのは受付済みだけというエラーになる",
			given: confirmedOrder, when: confirm,
			wantErr: order.ErrNotConfirmable, wantStatus: order.OrderStatusConfirmed,
		},
		{
			name:  "Given: 出荷済みの注文, When: 確定する, Then: 確定できるのは受付済みだけというエラーになる",
			given: shippedOrder, when: confirm,
			wantErr: order.ErrNotConfirmable, wantStatus: order.OrderStatusShipped,
		},
		{
			name:  "Given: キャンセルの注文, When: 確定する, Then: 確定できるのは受付済みだけというエラーになる",
			given: cancelledOrder, when: confirm,
			wantErr: order.ErrNotConfirmable, wantStatus: order.OrderStatusCancelled,
		},
		{
			name:  "Given: 受付済みの注文, When: キャンセルする, Then: キャンセルになり注文キャンセルが発行される",
			given: placedOrder, when: cancel,
			wantStatus: order.OrderStatusCancelled, wantEventType: order.EventTypeOrderCancelled,
		},
		{
			name:  "Given: 確定の注文, When: キャンセルする, Then: キャンセルになり注文キャンセルが発行される",
			given: confirmedOrder, when: cancel,
			wantStatus: order.OrderStatusCancelled, wantEventType: order.EventTypeOrderCancelled,
		},
		{
			name:  "Given: 出荷済みの注文, When: キャンセルする, Then: キャンセルできるのは出荷前だけというエラーになる",
			given: shippedOrder, when: cancel,
			wantErr: order.ErrNotCancellable, wantStatus: order.OrderStatusShipped,
		},
		{
			name:  "Given: キャンセルの注文, When: キャンセルする, Then: キャンセルできるのは出荷前だけというエラーになる",
			given: cancelledOrder, when: cancel,
			wantErr: order.ErrNotCancellable, wantStatus: order.OrderStatusCancelled,
		},
		{
			name:  "Given: 確定の注文, When: 出荷済みにする, Then: 出荷済みになり注文出荷が発行される",
			given: confirmedOrder, when: markShipped,
			wantStatus: order.OrderStatusShipped, wantEventType: order.EventTypeOrderShipped,
		},
		{
			name:  "Given: 受付済みの注文, When: 出荷済みにする, Then: 出荷済みにできるのは確定だけというエラーになる",
			given: placedOrder, when: markShipped,
			wantErr: order.ErrNotShippable, wantStatus: order.OrderStatusPlaced,
		},
		{
			name:  "Given: 出荷済みの注文, When: 出荷済みにする, Then: 出荷済みにできるのは確定だけというエラーになる",
			given: shippedOrder, when: markShipped,
			wantErr: order.ErrNotShippable, wantStatus: order.OrderStatusShipped,
		},
		{
			name:  "Given: キャンセルの注文, When: 出荷済みにする, Then: 出荷済みにできるのは確定だけというエラーになる",
			given: cancelledOrder, when: markShipped,
			wantErr: order.ErrNotShippable, wantStatus: order.OrderStatusCancelled,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := tt.given(t)
			beforeVersion := a.Version()
			beforeEvents := len(a.Events())

			err := tt.when(a)

			assert.Equal(t, tt.wantStatus, a.Status())
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.ErrorIs(t, err, order.ErrInvalidState)
				assert.Equal(t, beforeVersion, a.Version())
				assert.Len(t, a.Events(), beforeEvents)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, beforeVersion.Next(), a.Version())
			events := a.Events()
			require.Len(t, events, beforeEvents+1)
			last := events[len(events)-1]
			assert.Equal(t, tt.wantEventType, last.EventType())
			assert.Equal(t, a.ID(), last.OrderID())
			assert.Equal(t, event.NewMetadata(a.Version(), testCorrelationID, testNow), last.Metadata())
		})
	}
}

func TestAggregate_Confirm_OrderConfirmedの中身(t *testing.T) {
	a := placedOrder(t)

	require.NoError(t, a.Confirm(testCorrelationID, testNow))

	events := a.Events()
	confirmed, ok := events[len(events)-1].(order.OrderConfirmed)
	require.True(t, ok)
	assert.Equal(t, a.CustomerID(), confirmed.CustomerID())
	assert.Equal(t, order.PaymentMethodCreditCard, confirmed.PaymentMethod())
	assert.Equal(t, 2*1200+1*800, confirmed.TotalAmount())
}

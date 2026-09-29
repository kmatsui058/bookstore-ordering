package order_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/domain/order"
)

func TestRestoreFromEvents(t *testing.T) {
	tests := []struct {
		name    string
		given   func(t *testing.T) []order.Event
		wantErr error
	}{
		{
			name:  "Given: 注文済みから出荷までのイベントの列, When: 復元する, Then: 元の注文と同じ状態になる",
			given: func(t *testing.T) []order.Event { return shippedOrder(t).Events() },
		},
		{
			name:  "Given: 注文済みからキャンセルまでのイベントの列, When: 復元する, Then: 元の注文と同じ状態になる",
			given: func(t *testing.T) []order.Event { return cancelledOrder(t).Events() },
		},
		{
			name:    "Given: 空のイベントの列, When: 復元する, Then: イベントの列が壊れているというエラーになる",
			given:   func(*testing.T) []order.Event { return nil },
			wantErr: order.ErrBrokenEventStream,
		},
		{
			name:    "Given: 注文済みで始まらないイベントの列, When: 復元する, Then: イベントの列が壊れているというエラーになる",
			given:   func(t *testing.T) []order.Event { return shippedOrder(t).Events()[1:] },
			wantErr: order.ErrBrokenEventStream,
		},
		{
			name: "Given: バージョンが欠けたイベントの列, When: 復元する, Then: イベントの列が壊れているというエラーになる",
			given: func(t *testing.T) []order.Event {
				events := shippedOrder(t).Events()
				return []order.Event{events[0], events[2]}
			},
			wantErr: order.ErrBrokenEventStream,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events := tt.given(t)

			got, err := order.RestoreFromEvents(events, order.NewBus())

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			last := events[len(events)-1]
			placed, ok := events[0].(order.OrderPlaced)
			require.True(t, ok)
			assert.Equal(t, placed.OrderID(), got.ID())
			assert.Equal(t, placed.CustomerID(), got.CustomerID())
			assert.Equal(t, placed.Lines(), got.Lines())
			assert.Equal(t, placed.PaymentMethod(), got.PaymentMethod())
			assert.Equal(t, last.Metadata().Version(), got.Version())
			assert.Empty(t, got.Events())
		})
	}
}

func TestRestoreFromEvents_復元した注文は同じ状態になる(t *testing.T) {
	tests := []struct {
		name  string
		given func(t *testing.T) *order.Aggregate
	}{
		{name: "Given: 受付済みの注文のイベント, When: 復元する, Then: 受付済みになる", given: placedOrder},
		{name: "Given: 確定の注文のイベント, When: 復元する, Then: 確定になる", given: confirmedOrder},
		{name: "Given: 出荷済みの注文のイベント, When: 復元する, Then: 出荷済みになる", given: shippedOrder},
		{name: "Given: キャンセルの注文のイベント, When: 復元する, Then: キャンセルになる", given: cancelledOrder},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := tt.given(t)

			got, err := order.RestoreFromEvents(original.Events(), order.NewBus())

			require.NoError(t, err)
			assert.Equal(t, original.Status(), got.Status())
			assert.Equal(t, original.Version(), got.Version())
			assert.Equal(t, original.TotalAmount(), got.TotalAmount())
		})
	}
}

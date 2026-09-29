package inmemory_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/app/command"
	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/domain/order"
	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/infra/messaging/inprocess"
	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/infra/repository/inmemory"
	"github.com/kmatsui058/bookstore-ordering/internal/shared/domain/event"
)

const testCorrelationID = event.CorrelationID("corr-1")

// placeOrder は、ストアにつないだサービスで受付済みの注文を1件作る。
func placeOrder(t *testing.T, s *command.Service) *order.Aggregate {
	t.Helper()
	line, err := order.NewOrderLine("book-1", 3, 700)
	require.NoError(t, err)
	agg, err := s.PlaceOrder(context.Background(), command.PlaceOrderInput{
		CustomerID:    "customer-1",
		Lines:         []order.OrderLine{line},
		PaymentMethod: order.PaymentMethodBankTransfer,
		CorrelationID: testCorrelationID,
	})
	require.NoError(t, err)
	return agg
}

func TestStore_イベントの保存と復元の往復(t *testing.T) {
	store := inmemory.NewStore()
	tm := inmemory.NewTransactionManager(store)
	var published []order.EventType
	publisher := inprocess.NewPublisher(func(_ context.Context, e order.Event) error {
		published = append(published, e.EventType())
		return nil
	})
	s := command.NewService(tm, publisher)
	placed := placeOrder(t, s)
	_, err := s.ConfirmOrder(context.Background(), placed.ID(), testCorrelationID)
	require.NoError(t, err)
	_, err = s.MarkOrderShipped(context.Background(), placed.ID(), testCorrelationID)
	require.NoError(t, err)

	var restored *order.Aggregate
	err = tm.RunTransaction(context.Background(), func(ctx context.Context, tx command.TransactionalRepository) error {
		restored, err = tx.FindByID(ctx, placed.ID())
		return err
	})

	require.NoError(t, err)
	assert.Equal(t, placed.ID(), restored.ID())
	assert.Equal(t, placed.CustomerID(), restored.CustomerID())
	assert.Equal(t, placed.Lines(), restored.Lines())
	assert.Equal(t, placed.PaymentMethod(), restored.PaymentMethod())
	assert.Equal(t, order.OrderStatusShipped, restored.Status())
	assert.Equal(t, event.Version(3), restored.Version())
	assert.Equal(t, []order.EventType{order.EventTypeOrderPlaced, order.EventTypeOrderConfirmed, order.EventTypeOrderShipped}, published)

	snapshot, err := store.FindOrder(context.Background(), placed.ID())
	require.NoError(t, err)
	assert.Equal(t, order.OrderStatusShipped, snapshot.Status)
	assert.Equal(t, event.Version(3), snapshot.Version)
}

func TestTransactionManager_楽観的並行性制御(t *testing.T) {
	store := inmemory.NewStore()
	tm := inmemory.NewTransactionManager(store)
	s := command.NewService(tm, inprocess.NewPublisher())
	placed := placeOrder(t, s)

	err := tm.RunTransaction(context.Background(), func(ctx context.Context, tx command.TransactionalRepository) error {
		stale, err := tx.FindByID(ctx, placed.ID())
		require.NoError(t, err)
		_, err = s.ConfirmOrder(ctx, placed.ID(), testCorrelationID)
		require.NoError(t, err)
		require.NoError(t, stale.Cancel(testCorrelationID, time.Now()))
		return tx.SaveEvents(ctx, stale.Events())
	})

	require.ErrorIs(t, err, command.ErrConcurrencyConflict)
	snapshot, err := store.FindOrder(context.Background(), placed.ID())
	require.NoError(t, err)
	assert.Equal(t, order.OrderStatusConfirmed, snapshot.Status)
}

func TestStore_ない注文(t *testing.T) {
	tests := []struct {
		name string
		read func(store *inmemory.Store, tm *inmemory.TransactionManager, id order.OrderID) error
	}{
		{
			name: "Given: 空のストア, When: スナップショットを読む, Then: 注文が見つからないエラーになる",
			read: func(store *inmemory.Store, _ *inmemory.TransactionManager, id order.OrderID) error {
				_, err := store.FindOrder(context.Background(), id)
				return err
			},
		},
		{
			name: "Given: 空のストア, When: トランザクションの中で注文を読み込む, Then: 注文が見つからないエラーになる",
			read: func(_ *inmemory.Store, tm *inmemory.TransactionManager, id order.OrderID) error {
				return tm.RunTransaction(context.Background(), func(ctx context.Context, tx command.TransactionalRepository) error {
					_, err := tx.FindByID(ctx, id)
					return err
				})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := inmemory.NewStore()

			err := tt.read(store, inmemory.NewTransactionManager(store), order.NewOrderID())

			require.ErrorIs(t, err, order.ErrNotFound)
		})
	}
}

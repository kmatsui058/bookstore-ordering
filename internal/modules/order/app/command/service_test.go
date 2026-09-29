package command_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/app/command"
	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/app/command/mock_command"
	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/domain/order"
	"github.com/kmatsui058/bookstore-ordering/internal/shared/domain/event"
)

const testCorrelationID = event.CorrelationID("corr-1")

var errStore = errors.New("イベントストアに書けない")

// storedOrder は、指定した状態までコマンドを実行した注文を、読み込み直した（発行済みイベントが空の）集約として返す。
func storedOrder(t *testing.T, commands ...func(a *order.Aggregate) error) *order.Aggregate {
	t.Helper()
	line, err := order.NewOrderLine("book-1", 1, 1000)
	require.NoError(t, err)
	a, err := order.Place(order.NewOrderID(), "customer-1", []order.OrderLine{line}, order.PaymentMethodCreditCard, testCorrelationID, time.Now(), order.NewBus())
	require.NoError(t, err)
	for _, c := range commands {
		require.NoError(t, c(a))
	}
	restored, err := order.RestoreFromEvents(a.Events(), order.NewBus())
	require.NoError(t, err)
	return restored
}

func confirmed(a *order.Aggregate) error { return a.Confirm(testCorrelationID, time.Now()) }
func shipped(a *order.Aggregate) error   { return a.MarkShipped(testCorrelationID, time.Now()) }

// mocks は、サービスが要求するインターフェースのモック一式。
type mocks struct {
	tm        *mock_command.MockTransactionManager
	tx        *mock_command.MockTransactionalRepository
	publisher *mock_command.MockEventPublisher
}

func newMocks(t *testing.T) mocks {
	ctrl := gomock.NewController(t)
	m := mocks{
		tm:        mock_command.NewMockTransactionManager(ctrl),
		tx:        mock_command.NewMockTransactionalRepository(ctrl),
		publisher: mock_command.NewMockEventPublisher(ctrl),
	}
	m.tm.EXPECT().RunTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(context.Context, command.TransactionalRepository) error) error {
			return fn(ctx, m.tx)
		})
	return m
}

func TestService_CancelOrder(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, m mocks) order.OrderID
		wantErr    error
		wantStatus order.OrderStatus
	}{
		{
			name: "Given: 受付済みの注文, When: キャンセルする, Then: 注文キャンセルを保存してスナップショットを保存し、コミット後に公開する",
			setup: func(t *testing.T, m mocks) order.OrderID {
				agg := storedOrder(t)
				m.tx.EXPECT().FindByID(gomock.Any(), agg.ID()).Return(agg, nil)
				gomock.InOrder(
					m.tx.EXPECT().SaveEvents(gomock.Any(), gomock.Len(1)).Return(nil),
					m.tx.EXPECT().SaveSnapshot(gomock.Any(), agg).Return(nil),
					m.publisher.EXPECT().Publish(gomock.Any(), gomock.Len(1)).Return(nil),
				)
				return agg.ID()
			},
			wantStatus: order.OrderStatusCancelled,
		},
		{
			name: "Given: 確定の注文, When: キャンセルする, Then: キャンセルになる",
			setup: func(t *testing.T, m mocks) order.OrderID {
				agg := storedOrder(t, confirmed)
				m.tx.EXPECT().FindByID(gomock.Any(), agg.ID()).Return(agg, nil)
				m.tx.EXPECT().SaveEvents(gomock.Any(), gomock.Len(1)).Return(nil)
				m.tx.EXPECT().SaveSnapshot(gomock.Any(), agg).Return(nil)
				m.publisher.EXPECT().Publish(gomock.Any(), gomock.Len(1)).Return(nil)
				return agg.ID()
			},
			wantStatus: order.OrderStatusCancelled,
		},
		{
			name: "Given: 出荷済みの注文, When: キャンセルする, Then: キャンセルできないエラーになり、何も保存も公開もしない",
			setup: func(t *testing.T, m mocks) order.OrderID {
				agg := storedOrder(t, confirmed, shipped)
				m.tx.EXPECT().FindByID(gomock.Any(), agg.ID()).Return(agg, nil)
				return agg.ID()
			},
			wantErr: order.ErrNotCancellable,
		},
		{
			name: "Given: ない注文ID, When: キャンセルする, Then: 注文が見つからないエラーになる",
			setup: func(t *testing.T, m mocks) order.OrderID {
				m.tx.EXPECT().FindByID(gomock.Any(), gomock.Any()).Return(nil, order.ErrNotFound)
				return order.NewOrderID()
			},
			wantErr: order.ErrNotFound,
		},
		{
			name: "Given: イベントの保存に失敗する, When: キャンセルする, Then: そのエラーになり、スナップショットの保存も公開もしない",
			setup: func(t *testing.T, m mocks) order.OrderID {
				agg := storedOrder(t)
				m.tx.EXPECT().FindByID(gomock.Any(), agg.ID()).Return(agg, nil)
				m.tx.EXPECT().SaveEvents(gomock.Any(), gomock.Any()).Return(errStore)
				return agg.ID()
			},
			wantErr: errStore,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newMocks(t)
			orderID := tt.setup(t, m)
			s := command.NewService(m.tm, m.publisher)

			got, err := s.CancelOrder(context.Background(), orderID, testCorrelationID)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, got.Status())
		})
	}
}

func TestService_状態を変えるコマンド(t *testing.T) {
	tests := []struct {
		name       string
		given      []func(a *order.Aggregate) error
		when       func(s *command.Service, id order.OrderID) (*order.Aggregate, error)
		wantStatus order.OrderStatus
		wantEvent  order.EventType
	}{
		{
			name:  "Given: 受付済みの注文, When: 確定する, Then: 確定になり注文確定を公開する",
			given: nil,
			when: func(s *command.Service, id order.OrderID) (*order.Aggregate, error) {
				return s.ConfirmOrder(context.Background(), id, testCorrelationID)
			},
			wantStatus: order.OrderStatusConfirmed,
			wantEvent:  order.EventTypeOrderConfirmed,
		},
		{
			name:  "Given: 確定の注文, When: 出荷済みにする, Then: 出荷済みになり注文出荷を公開する",
			given: []func(a *order.Aggregate) error{confirmed},
			when: func(s *command.Service, id order.OrderID) (*order.Aggregate, error) {
				return s.MarkOrderShipped(context.Background(), id, testCorrelationID)
			},
			wantStatus: order.OrderStatusShipped,
			wantEvent:  order.EventTypeOrderShipped,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newMocks(t)
			agg := storedOrder(t, tt.given...)
			m.tx.EXPECT().FindByID(gomock.Any(), agg.ID()).Return(agg, nil)
			m.tx.EXPECT().SaveEvents(gomock.Any(), gomock.Len(1)).Return(nil)
			m.tx.EXPECT().SaveSnapshot(gomock.Any(), agg).Return(nil)
			var published []order.Event
			m.publisher.EXPECT().Publish(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, events []order.Event) error {
					published = events
					return nil
				})
			s := command.NewService(m.tm, m.publisher)

			got, err := tt.when(s, agg.ID())

			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, got.Status())
			require.Len(t, published, 1)
			assert.Equal(t, tt.wantEvent, published[0].EventType())
		})
	}
}

func TestService_PlaceOrder(t *testing.T) {
	line, err := order.NewOrderLine("book-1", 2, 500)
	require.NoError(t, err)
	tests := []struct {
		name    string
		input   command.PlaceOrderInput
		expect  func(m mocks)
		wantErr error
	}{
		{
			name:  "Given: 正しい注文の入力, When: 注文する, Then: 注文済みを保存して公開し、受付済みの注文を返す",
			input: command.PlaceOrderInput{CustomerID: "customer-1", Lines: []order.OrderLine{line}, PaymentMethod: order.PaymentMethodCashOnDelivery, CorrelationID: testCorrelationID},
			expect: func(m mocks) {
				m.tx.EXPECT().SaveEvents(gomock.Any(), gomock.Len(1)).Return(nil)
				m.tx.EXPECT().SaveSnapshot(gomock.Any(), gomock.Any()).Return(nil)
				m.publisher.EXPECT().Publish(gomock.Any(), gomock.Len(1)).Return(nil)
			},
		},
		{
			name:    "Given: 注文明細のない入力, When: 注文する, Then: 注文明細は1件以上必要というエラーになり、何も保存しない",
			input:   command.PlaceOrderInput{CustomerID: "customer-1", PaymentMethod: order.PaymentMethodCreditCard, CorrelationID: testCorrelationID},
			expect:  func(mocks) {},
			wantErr: order.ErrNoLines,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newMocks(t)
			tt.expect(m)
			s := command.NewService(m.tm, m.publisher)

			got, err := s.PlaceOrder(context.Background(), tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, order.OrderStatusPlaced, got.Status())
			assert.Equal(t, tt.input.CustomerID, got.CustomerID())
		})
	}
}

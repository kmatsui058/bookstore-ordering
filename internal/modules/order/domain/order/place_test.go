package order_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/domain/order"
	"github.com/kmatsui058/bookstore-ordering/internal/shared/domain/event"
	"github.com/kmatsui058/bookstore-ordering/internal/shared/domain/id"
)

func TestPlace(t *testing.T) {
	validLines := func(t *testing.T) []order.OrderLine {
		return []order.OrderLine{newLine(t, "book-1", 2, 1200)}
	}
	tests := []struct {
		name          string
		customerID    id.CustomerID
		lines         func(t *testing.T) []order.OrderLine
		paymentMethod order.PaymentMethod
		wantErr       error
	}{
		{
			name:          "Given: 顧客ID・注文明細1件・支払方法がそろっている, When: 注文する, Then: 受付済みの注文ができ注文済みが発行される",
			customerID:    "customer-1",
			lines:         validLines,
			paymentMethod: order.PaymentMethodBankTransfer,
		},
		{
			name:          "Given: 注文明細が0件, When: 注文する, Then: 注文明細は1件以上必要というエラーになる",
			customerID:    "customer-1",
			lines:         func(*testing.T) []order.OrderLine { return nil },
			paymentMethod: order.PaymentMethodCreditCard,
			wantErr:       order.ErrNoLines,
		},
		{
			name:          "Given: NewOrderLine を通さないゼロ値の注文明細, When: 注文する, Then: 注文明細のルールを満たさないというエラーになる",
			customerID:    "customer-1",
			lines:         func(*testing.T) []order.OrderLine { return []order.OrderLine{{}} },
			paymentMethod: order.PaymentMethodCreditCard,
			wantErr:       order.ErrMissingBookID,
		},
		{
			name:          "Given: 顧客ID が空, When: 注文する, Then: 顧客ID は必須というエラーになる",
			customerID:    "",
			lines:         validLines,
			paymentMethod: order.PaymentMethodCreditCard,
			wantErr:       order.ErrMissingCustomerID,
		},
		{
			name:          "Given: 定義にない支払方法, When: 注文する, Then: 支払方法が正しくないというエラーになる",
			customerID:    "customer-1",
			lines:         validLines,
			paymentMethod: order.PaymentMethod("POINT"),
			wantErr:       order.ErrInvalidPaymentMethod,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderID := order.NewOrderID()
			lines := tt.lines(t)

			got, err := order.Place(orderID, tt.customerID, lines, tt.paymentMethod, testCorrelationID, testNow, order.NewBus())

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.ErrorIs(t, err, order.ErrInvalidArgument)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, orderID, got.ID())
			assert.Equal(t, tt.customerID, got.CustomerID())
			assert.Equal(t, lines, got.Lines())
			assert.Equal(t, tt.paymentMethod, got.PaymentMethod())
			assert.Equal(t, order.OrderStatusPlaced, got.Status())
			assert.Equal(t, event.Version(1), got.Version())
			require.Len(t, got.Events(), 1)
			placed, ok := got.Events()[0].(order.OrderPlaced)
			require.True(t, ok)
			assert.Equal(t, orderID, placed.OrderID())
			assert.Equal(t, event.NewMetadata(1, testCorrelationID, testNow), placed.Metadata())
		})
	}
}

func TestNewOrderLine(t *testing.T) {
	tests := []struct {
		name      string
		bookID    id.BookID
		quantity  int
		unitPrice int
		wantErr   error
	}{
		{name: "Given: 数量1・単価0, When: 注文明細を作る, Then: 作れる", bookID: "book-1", quantity: 1, unitPrice: 0},
		{name: "Given: 数量0, When: 注文明細を作る, Then: 数量は1以上というエラーになる", bookID: "book-1", quantity: 0, unitPrice: 100, wantErr: order.ErrInvalidQuantity},
		{name: "Given: 数量が負, When: 注文明細を作る, Then: 数量は1以上というエラーになる", bookID: "book-1", quantity: -1, unitPrice: 100, wantErr: order.ErrInvalidQuantity},
		{name: "Given: 単価が負, When: 注文明細を作る, Then: 単価は0以上というエラーになる", bookID: "book-1", quantity: 1, unitPrice: -1, wantErr: order.ErrInvalidUnitPrice},
		{name: "Given: 書籍ID が空, When: 注文明細を作る, Then: 書籍ID は必須というエラーになる", bookID: "", quantity: 1, unitPrice: 100, wantErr: order.ErrMissingBookID},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := order.NewOrderLine(tt.bookID, tt.quantity, tt.unitPrice)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.bookID, got.BookID())
			assert.Equal(t, tt.quantity, got.Quantity())
			assert.Equal(t, tt.unitPrice, got.UnitPrice())
		})
	}
}

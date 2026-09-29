package order_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/domain/order"
)

func TestPlace_すべての支払方法で注文できる(t *testing.T) {
	for _, pm := range order.AllPaymentMethods() {
		t.Run("Given: 支払方法 "+pm.String()+", When: 注文する, Then: 受付済みになる", func(t *testing.T) {
			a, err := order.Place(order.NewOrderID(), "customer-1", []order.OrderLine{newLine(t, "book-1", 1, 100)}, pm, testCorrelationID, testNow, order.NewBus())

			require.NoError(t, err)
			assert.Equal(t, pm, a.PaymentMethod())
		})
	}
}

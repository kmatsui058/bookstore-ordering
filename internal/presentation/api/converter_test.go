package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/domain/order"
)

func TestToAPIOrderStatus_ドメインの値は契約の値と一致する(t *testing.T) {
	for _, s := range order.AllOrderStatuses() {
		t.Run("Given: 注文ステータス "+s.String()+", When: API の値に変換する, Then: 契約に定義された同じ値になる", func(t *testing.T) {
			got := toAPIOrderStatus(s)

			assert.True(t, got.Valid())
			assert.Equal(t, s.String(), string(got))
		})
	}
}

func TestPaymentMethod_ドメインと契約の値が往復する(t *testing.T) {
	for _, p := range order.AllPaymentMethods() {
		t.Run("Given: 支払方法 "+p.String()+", When: API の値に変換して戻す, Then: 契約に定義された同じ値で元に戻る", func(t *testing.T) {
			apiValue := toAPIPaymentMethod(p)

			back, err := toDomainPaymentMethod(apiValue)

			require.NoError(t, err)
			assert.True(t, apiValue.Valid())
			assert.Equal(t, p.String(), string(apiValue))
			assert.Equal(t, p, back)
		})
	}
}

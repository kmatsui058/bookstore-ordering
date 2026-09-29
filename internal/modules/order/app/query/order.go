// Package query は、注文を読み取るユースケースを実行するアプリケーション層のパッケージ。
// 集約をイベントから復元せず、コマンドの側が保存したスナップショットを読む。
package query

import (
	"context"

	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/domain/order"
	"github.com/kmatsui058/bookstore-ordering/internal/shared/domain/event"
	"github.com/kmatsui058/bookstore-ordering/internal/shared/domain/id"
)

// Order は、注文の読み取り用のモデル（スナップショット）。状態を変える振る舞いは持たない。
type Order struct {
	OrderID       order.OrderID
	CustomerID    id.CustomerID
	Lines         []OrderLine
	PaymentMethod order.PaymentMethod
	Status        order.OrderStatus
	Version       event.Version
}

// OrderLine は、注文明細の読み取り用のモデル。
type OrderLine struct {
	BookID    id.BookID
	Quantity  int
	UnitPrice int
}

// NewOrder は、集約のいまの状態から読み取り用のモデルを作る。
func NewOrder(agg *order.Aggregate) Order {
	lines := make([]OrderLine, 0, len(agg.Lines()))
	for _, l := range agg.Lines() {
		lines = append(lines, OrderLine{BookID: l.BookID(), Quantity: l.Quantity(), UnitPrice: l.UnitPrice()})
	}
	return Order{
		OrderID:       agg.ID(),
		CustomerID:    agg.CustomerID(),
		Lines:         lines,
		PaymentMethod: agg.PaymentMethod(),
		Status:        agg.Status(),
		Version:       agg.Version(),
	}
}

// Service は、注文を読み取るユースケースを実行するサービス。
type Service struct {
	reader SnapshotReader
}

// NewService は、注文を読み取るサービスを作る。
func NewService(reader SnapshotReader) *Service {
	return &Service{reader: reader}
}

// GetOrder は、注文ID で注文を1件読む。
func (s *Service) GetOrder(ctx context.Context, orderID order.OrderID) (Order, error) {
	return s.reader.FindOrder(ctx, orderID)
}

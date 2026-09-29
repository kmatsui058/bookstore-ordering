package command

import (
	"context"
	"time"

	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/domain/order"
	"github.com/kmatsui058/bookstore-ordering/internal/shared/domain/event"
	"github.com/kmatsui058/bookstore-ordering/internal/shared/domain/id"
)

// Service は、注文のコマンド（注文する・確定する・キャンセルする・出荷済みにする）のユースケースを実行するアプリケーション層のサービス。
// 集約の読み込み・コマンドの実行・永続化を1つのトランザクションで行い、コミットしたあとでイベントを公開する。
type Service struct {
	tm        TransactionManager
	publisher EventPublisher
	now       func() time.Time
}

// NewService は、注文のコマンドのサービスを作る。
func NewService(tm TransactionManager, publisher EventPublisher) *Service {
	return &Service{tm: tm, publisher: publisher, now: time.Now}
}

// PlaceOrderInput は、ユースケース「注文する」の入力。
type PlaceOrderInput struct {
	CustomerID    id.CustomerID
	Lines         []order.OrderLine
	PaymentMethod order.PaymentMethod
	CorrelationID event.CorrelationID
}

// PlaceOrder は、新しい注文ID を採番して注文を受け付ける。
func (s *Service) PlaceOrder(ctx context.Context, in PlaceOrderInput) (*order.Aggregate, error) {
	var placed *order.Aggregate
	err := s.tm.RunTransaction(ctx, func(ctx context.Context, tx TransactionalRepository) error {
		agg, err := order.Place(order.NewOrderID(), in.CustomerID, in.Lines, in.PaymentMethod, in.CorrelationID, s.now(), order.NewBus())
		if err != nil {
			return err
		}
		if err := persist(ctx, tx, agg); err != nil {
			return err
		}
		placed = agg
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.publish(ctx, placed)
}

// ConfirmOrder は、受付済みの注文を確定する。
func (s *Service) ConfirmOrder(ctx context.Context, orderID order.OrderID, correlationID event.CorrelationID) (*order.Aggregate, error) {
	return s.change(ctx, orderID, func(agg *order.Aggregate) error {
		return agg.Confirm(correlationID, s.now())
	})
}

// CancelOrder は、出荷前の注文をキャンセルする。
func (s *Service) CancelOrder(ctx context.Context, orderID order.OrderID, correlationID event.CorrelationID) (*order.Aggregate, error) {
	return s.change(ctx, orderID, func(agg *order.Aggregate) error {
		return agg.Cancel(correlationID, s.now())
	})
}

// MarkOrderShipped は、確定した注文を出荷済みにする。
func (s *Service) MarkOrderShipped(ctx context.Context, orderID order.OrderID, correlationID event.CorrelationID) (*order.Aggregate, error) {
	return s.change(ctx, orderID, func(agg *order.Aggregate) error {
		return agg.MarkShipped(correlationID, s.now())
	})
}

// change は、既存の注文を読み込んでコマンドを実行し、永続化してからイベントを公開する。
func (s *Service) change(ctx context.Context, orderID order.OrderID, execute func(agg *order.Aggregate) error) (*order.Aggregate, error) {
	var changed *order.Aggregate
	err := s.tm.RunTransaction(ctx, func(ctx context.Context, tx TransactionalRepository) error {
		agg, err := tx.FindByID(ctx, orderID)
		if err != nil {
			return err
		}
		if err := execute(agg); err != nil {
			return err
		}
		if err := persist(ctx, tx, agg); err != nil {
			return err
		}
		changed = agg
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.publish(ctx, changed)
}

// publish は、コミットしたあとで集約のイベントを公開する。
// TODO: 公開に失敗したイベントを再送する仕組み（トランザクションアウトボックスなど）がない。実際のメッセージ基盤につなぐときに入れる。
func (s *Service) publish(ctx context.Context, agg *order.Aggregate) (*order.Aggregate, error) {
	if err := s.publisher.Publish(ctx, agg.Events()); err != nil {
		return nil, err
	}
	return agg, nil
}

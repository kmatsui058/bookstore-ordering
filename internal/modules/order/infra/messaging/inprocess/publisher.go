// Package inprocess は、コミットした注文のドメインイベントを同じプロセスの中の購読者に渡すインフラ層の公開者。
// 実際のメッセージ基盤の代わりに使うサンプルの実装。
package inprocess

import (
	"context"

	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/app/command"
	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/domain/order"
)

// Subscriber は、公開されたドメインイベントを受け取る購読者。
type Subscriber func(ctx context.Context, e order.Event) error

// Publisher は、イベントを発行した順に、登録した購読者へ同期的に渡す。
type Publisher struct {
	subscribers []Subscriber
}

var _ command.EventPublisher = (*Publisher)(nil)

// NewPublisher は、購読者を登録した公開者を作る。
func NewPublisher(subscribers ...Subscriber) *Publisher {
	return &Publisher{subscribers: subscribers}
}

// Publish は、各イベントをすべての購読者に渡す。購読者が失敗したら、そこで止めてエラーを返す。
func (p *Publisher) Publish(ctx context.Context, events []order.Event) error {
	for _, e := range events {
		for _, s := range p.subscribers {
			if err := s(ctx, e); err != nil {
				return err
			}
		}
	}
	return nil
}

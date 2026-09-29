//go:generate go tool mockgen -source=$GOFILE -destination=mock_$GOPACKAGE/mock_$GOFILE -package=mock_$GOPACKAGE
package command

import (
	"context"
	"errors"

	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/domain/order"
)

// ErrConcurrencyConflict は、読み込んだあとに同じ注文が別の処理で変わり、同じバージョンのイベントを保存できなかったことを表す。
// インフラ層のリポジトリは、楽観的並行性制御で衝突を見つけたらこのエラーを返す。
var ErrConcurrencyConflict = errors.New("同じ注文が同時に変更されました")

// TransactionManager は、ユースケースの読み込みから保存までを1つのトランザクションで実行する。アプリケーション層が要求し、インフラ層が実装する。
type TransactionManager interface {
	// RunTransaction は、fn をトランザクションの中で実行し、fn がエラーを返さなければコミットする。
	RunTransaction(ctx context.Context, fn func(ctx context.Context, tx TransactionalRepository) error) error
}

// TransactionalRepository は、トランザクションの中で注文を読み書きするリポジトリ。
type TransactionalRepository interface {
	// FindByID は、注文のイベントの列から注文を復元する。ないときは order.ErrNotFound を返す。
	FindByID(ctx context.Context, orderID order.OrderID) (*order.Aggregate, error)
	// SaveEvents は、ドメインイベントをイベントストアに追記する。同じバージョンがすでにあれば ErrConcurrencyConflict を返す。
	SaveEvents(ctx context.Context, events []order.Event) error
	// SaveSnapshot は、読み取り用に、集約のいまの状態をスナップショットとして保存する。
	SaveSnapshot(ctx context.Context, agg *order.Aggregate) error
}

// EventPublisher は、コミットしたドメインイベントをモジュールやコンテキストの外に公開する。
type EventPublisher interface {
	// Publish は、イベントを発行した順に公開する。
	Publish(ctx context.Context, events []order.Event) error
}

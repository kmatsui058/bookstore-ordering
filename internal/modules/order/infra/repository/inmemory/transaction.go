package inmemory

import (
	"context"

	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/app/command"
	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/app/query"
	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/domain/order"
)

// TransactionManager は、ストアへの書き込みをトランザクションの単位にまとめる。
// トランザクションの中では書き込みを積むだけにし、fn が成功したときにだけコミットする（ロックはコミットのときだけ取る）。
type TransactionManager struct {
	store *Store
}

var _ command.TransactionManager = (*TransactionManager)(nil)

// NewTransactionManager は、ストアに書き込むトランザクションマネージャーを作る。
func NewTransactionManager(store *Store) *TransactionManager {
	return &TransactionManager{store: store}
}

// RunTransaction は、fn をトランザクションの中で実行し、成功したらコミットする。失敗したら積んだ書き込みを捨てる。
func (m *TransactionManager) RunTransaction(ctx context.Context, fn func(ctx context.Context, tx command.TransactionalRepository) error) error {
	tx := &transactionalRepository{store: m.store}
	if err := fn(ctx, tx); err != nil {
		return err
	}
	return m.store.commit(tx.events, tx.snapshots)
}

// transactionalRepository は、1つのトランザクションの中で注文を読み、書き込みを積むリポジトリ。
type transactionalRepository struct {
	store     *Store
	events    []order.Event
	snapshots []query.Order
}

var _ command.TransactionalRepository = (*transactionalRepository)(nil)

// FindByID は、コミット済みのイベントの列から注文を復元する。ないときは order.ErrNotFound を返す。
func (r *transactionalRepository) FindByID(_ context.Context, orderID order.OrderID) (*order.Aggregate, error) {
	events := r.store.stream(orderID)
	if len(events) == 0 {
		return nil, order.ErrNotFound
	}
	return order.RestoreFromEvents(events, order.NewBus())
}

// SaveEvents は、コミットのときに追記するイベントを積む。
func (r *transactionalRepository) SaveEvents(_ context.Context, events []order.Event) error {
	r.events = append(r.events, events...)
	return nil
}

// SaveSnapshot は、コミットのときに保存するスナップショットを積む。
func (r *transactionalRepository) SaveSnapshot(_ context.Context, agg *order.Aggregate) error {
	r.snapshots = append(r.snapshots, query.NewOrder(agg))
	return nil
}

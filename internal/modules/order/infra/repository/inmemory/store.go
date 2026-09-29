// Package inmemory は、注文のイベントストアとスナップショットをメモリに持つインフラ層のリポジトリ。
// サンプルとテストのための実装で、プロセスを止めると中身は消える。
package inmemory

import (
	"context"
	"fmt"
	"sync"

	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/app/command"
	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/app/query"
	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/domain/order"
)

// Store は、注文ごとのイベントの列（正本）と、読み取り用のスナップショットを持つ。
// 書き込みはトランザクションのコミットでだけ行い、同じバージョンのイベントがすでにあれば衝突として拒む。
type Store struct {
	mu        sync.RWMutex
	streams   map[order.OrderID][]order.Event
	snapshots map[order.OrderID]query.Order
}

var _ query.SnapshotReader = (*Store)(nil)

// NewStore は、空のストアを作る。
func NewStore() *Store {
	return &Store{
		streams:   map[order.OrderID][]order.Event{},
		snapshots: map[order.OrderID]query.Order{},
	}
}

// FindOrder は、注文のスナップショットを読む。ないときは order.ErrNotFound を返す。
func (s *Store) FindOrder(_ context.Context, orderID order.OrderID) (query.Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	snapshot, ok := s.snapshots[orderID]
	if !ok {
		return query.Order{}, order.ErrNotFound
	}
	return snapshot, nil
}

// stream は、注文のイベントの列の複製を返す。
func (s *Store) stream(orderID order.OrderID) []order.Event {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]order.Event(nil), s.streams[orderID]...)
}

// commit は、トランザクションで積んだイベントとスナップショットをまとめて書き込む。
// どれか1つでもバージョンが衝突したら、何も書き込まずに command.ErrConcurrencyConflict を返す。
func (s *Store) commit(events []order.Event, snapshots []query.Order) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	lengths := map[order.OrderID]int{}
	for _, e := range events {
		if _, ok := lengths[e.OrderID()]; !ok {
			lengths[e.OrderID()] = len(s.streams[e.OrderID()])
		}
		want := lengths[e.OrderID()] + 1
		if e.Metadata().Version().Int() != want {
			return fmt.Errorf("%w: 注文 %s のバージョン %d を書こうとしましたが、次は %d です", command.ErrConcurrencyConflict, e.OrderID(), e.Metadata().Version(), want)
		}
		lengths[e.OrderID()] = want
	}
	for _, e := range events {
		s.streams[e.OrderID()] = append(s.streams[e.OrderID()], e)
	}
	for _, snapshot := range snapshots {
		s.snapshots[snapshot.OrderID] = snapshot
	}
	return nil
}

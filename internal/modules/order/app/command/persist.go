package command

import (
	"context"

	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/domain/order"
)

// persist は、集約が発行したイベントを追記し、そのあとで読み取り用のスナップショットを保存する。
// イベントが正本なので、イベントの追記に失敗したらスナップショットは保存しない。
func persist(ctx context.Context, tx TransactionalRepository, agg *order.Aggregate) error {
	if err := tx.SaveEvents(ctx, agg.Events()); err != nil {
		return err
	}
	return tx.SaveSnapshot(ctx, agg)
}

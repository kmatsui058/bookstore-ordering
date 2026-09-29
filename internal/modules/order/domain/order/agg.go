package order

import (
	"time"

	"github.com/kmatsui058/bookstore-ordering/internal/shared/domain/event"
	"github.com/kmatsui058/bookstore-ordering/internal/shared/domain/id"
)

// Aggregate は、顧客が書籍を買うために出した1回の申し込み（集約ルート「注文」。LikeC4 ビュー bkst_ordr_ordm_ord）。
//
// 不変条件:
//   - 注文明細は1件以上あり、各明細の数量は1以上である
//   - 状態は 受付済み → 確定 → 出荷済み、または出荷前（受付済み・確定）→ キャンセル の順にしか変わらない
//
// 状態はイベントを当てはめることでだけ変わり、version はそのたびに進む（楽観的並行性制御に使う）。
type Aggregate struct {
	orderID       OrderID
	customerID    id.CustomerID
	lines         []OrderLine
	paymentMethod PaymentMethod
	status        OrderStatus
	version       event.Version

	bus *Bus
}

// ID は、注文ID を返す。
func (a *Aggregate) ID() OrderID {
	return a.orderID
}

// CustomerID は、注文した顧客の顧客ID を返す。
func (a *Aggregate) CustomerID() id.CustomerID {
	return a.customerID
}

// Lines は、注文明細一覧の複製を返す。
func (a *Aggregate) Lines() []OrderLine {
	return append([]OrderLine(nil), a.lines...)
}

// PaymentMethod は、支払方法を返す。
func (a *Aggregate) PaymentMethod() PaymentMethod {
	return a.paymentMethod
}

// Status は、注文ステータスを返す。
func (a *Aggregate) Status() OrderStatus {
	return a.status
}

// Version は、最後に当てはめたイベントのバージョンを返す。
func (a *Aggregate) Version() event.Version {
	return a.version
}

// TotalAmount は、注文明細の単価 × 数量の合計（円）を返す。
func (a *Aggregate) TotalAmount() int {
	total := 0
	for _, l := range a.lines {
		total += l.Subtotal()
	}
	return total
}

// Events は、この集約を読み込んでから発行したドメインイベントを発行した順に返す。
func (a *Aggregate) Events() []Event {
	return a.bus.Events()
}

// nextMetadata は、次に発行するイベントのメタデータを作る。
func (a *Aggregate) nextMetadata(correlationID event.CorrelationID, occurredAt time.Time) event.Metadata {
	return event.NewMetadata(a.version.Next(), correlationID, occurredAt)
}

// raise は、イベントを当てはめてバージョンを進め、イベントバスに発行する。
func (a *Aggregate) raise(e Event) {
	a.applyEvent(e)
	a.bus.Publish(e)
}

// applyEvent は、イベントを当てはめてバージョンをイベントのバージョンに合わせる。
func (a *Aggregate) applyEvent(e Event) {
	e.apply(a)
	a.version = e.Metadata().Version()
}

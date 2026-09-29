package order

import "github.com/kmatsui058/bookstore-ordering/internal/shared/domain/event"

// EventType は、注文のドメインイベントの種類を表す列挙型。値はモデルのイベントの英名と同じにする。
type EventType string

const (
	// EventTypeOrderPlaced は、注文済み（OrderPlaced）。
	EventTypeOrderPlaced EventType = "OrderPlaced"
	// EventTypeOrderConfirmed は、注文確定（OrderConfirmed）。
	EventTypeOrderConfirmed EventType = "OrderConfirmed"
	// EventTypeOrderCancelled は、注文キャンセル（OrderCancelled）。
	EventTypeOrderCancelled EventType = "OrderCancelled"
	// EventTypeOrderShipped は、注文出荷（OrderShipped）。
	EventTypeOrderShipped EventType = "OrderShipped"
)

// AllEventTypes は、注文のドメインイベントの種類をすべて返す。イベントを足したらここにも足す。
func AllEventTypes() []EventType {
	return []EventType{EventTypeOrderPlaced, EventTypeOrderConfirmed, EventTypeOrderCancelled, EventTypeOrderShipped}
}

// Event は、注文のすべてのドメインイベントが満たす契約。
// apply を非公開にし、このパッケージの外からイベントを作って集約の状態を変えられないようにする。
type Event interface {
	// EventType は、イベントの種類を返す。
	EventType() EventType
	// OrderID は、イベントが起きた注文の注文ID を返す。
	OrderID() OrderID
	// Metadata は、イベントのバージョン・相関 ID・発生日時を返す。
	Metadata() event.Metadata
	// apply は、イベントが表す変化を集約の状態に当てはめる。
	apply(a *Aggregate)
}

// orderEvent は、注文のドメインイベントが共通に持つ注文ID とメタデータ。各イベントに埋め込む。
type orderEvent struct {
	orderID  OrderID
	metadata event.Metadata
}

// OrderID は、イベントが起きた注文の注文ID を返す。
func (e orderEvent) OrderID() OrderID {
	return e.orderID
}

// Metadata は、イベントのメタデータを返す。
func (e orderEvent) Metadata() event.Metadata {
	return e.metadata
}

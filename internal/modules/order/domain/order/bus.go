package order

// Bus は、集約が発行したドメインイベントを発行した順に記録するイベントバス。
// 記録したイベントは、アプリケーション層が永続化と外部への公開に使う。
type Bus struct {
	events []Event
}

// NewBus は、空のイベントバスを作る。
func NewBus() *Bus {
	return &Bus{}
}

// Publish は、イベントを記録する。
func (b *Bus) Publish(e Event) {
	b.events = append(b.events, e)
}

// Events は、記録したイベントの複製を発行した順に返す。
func (b *Bus) Events() []Event {
	return append([]Event(nil), b.events...)
}

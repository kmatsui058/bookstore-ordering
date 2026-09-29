package order

// Handler は、注文のドメインイベントを同じモジュールの中で受け取る購読者。
type Handler func(e Event) error

// Bus は、集約が発行したドメインイベントを記録し、同じモジュールの中の購読者に即座に伝えるイベントバス。
// 記録したイベントは、アプリケーション層が永続化と外部への公開に使う。
type Bus struct {
	handlers []Handler
	events   []Event
}

// NewBus は、購読者を持つイベントバスを作る。
func NewBus(handlers ...Handler) *Bus {
	return &Bus{handlers: handlers}
}

// Publish は、イベントを記録し、購読者に順に伝える。購読者が失敗したら、そのエラーを返す。
func (b *Bus) Publish(e Event) error {
	b.events = append(b.events, e)
	for _, h := range b.handlers {
		if err := h(e); err != nil {
			return err
		}
	}
	return nil
}

// Events は、記録したイベントの複製を発行した順に返す。
func (b *Bus) Events() []Event {
	return append([]Event(nil), b.events...)
}

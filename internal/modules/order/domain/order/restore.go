package order

import "fmt"

// RestoreFromEvents は、イベントの列から注文を復元する。インフラ層のリポジトリが集約を読み込むときに使う。
// 列は注文済み（OrderPlaced）で始まり、バージョンが 1 から欠けずに続いていなければならない。
// 復元ではイベントを発行しないので、渡したイベントバスは空のままになる。
func RestoreFromEvents(events []Event, bus *Bus) (*Aggregate, error) {
	if len(events) == 0 {
		return nil, fmt.Errorf("%w: イベントがありません", ErrBrokenEventStream)
	}
	if events[0].EventType() != EventTypeOrderPlaced {
		return nil, fmt.Errorf("%w: 最初のイベントが注文済みではありません", ErrBrokenEventStream)
	}
	a := &Aggregate{bus: bus}
	for _, e := range events {
		if e.Metadata().Version() != a.version.Next() {
			return nil, fmt.Errorf("%w: バージョン %d の次に %d が来ました", ErrBrokenEventStream, a.version, e.Metadata().Version())
		}
		a.applyEvent(e)
	}
	return a, nil
}

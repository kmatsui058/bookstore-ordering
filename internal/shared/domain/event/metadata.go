// Package event は、注文コンテキストのすべてのドメインイベントに共通するメタデータを表すドメイン層のパッケージ。
package event

import "time"

// Version は、集約のイベントストリームの中でのイベントの通し番号。1 から始まり、楽観的並行性制御に使う。
type Version int

// Next は、このバージョンの次のバージョンを返す。
func (v Version) Next() Version {
	return v + 1
}

// Int は、バージョンを int として返す。
func (v Version) Int() int {
	return int(v)
}

// CorrelationID は、1 つの業務の流れ（サガ）の中で起きたイベントを結び付ける識別子。
type CorrelationID string

// String は、相関 ID を文字列として返す。
func (c CorrelationID) String() string {
	return string(c)
}

// Metadata は、ドメインイベントに共通する付随情報（バージョン・相関 ID・発生日時）を持つ値オブジェクト。
type Metadata struct {
	version       Version
	correlationID CorrelationID
	occurredAt    time.Time
}

// NewMetadata は、イベントのメタデータを作る。
func NewMetadata(version Version, correlationID CorrelationID, occurredAt time.Time) Metadata {
	return Metadata{
		version:       version,
		correlationID: correlationID,
		occurredAt:    occurredAt,
	}
}

// Version は、イベントのバージョンを返す。
func (m Metadata) Version() Version {
	return m.version
}

// CorrelationID は、イベントの相関 ID を返す。
func (m Metadata) CorrelationID() CorrelationID {
	return m.correlationID
}

// OccurredAt は、イベントが起きた日時を返す。
func (m Metadata) OccurredAt() time.Time {
	return m.occurredAt
}

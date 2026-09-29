// Package id は、注文コンテキストの共通語彙（LikeC4 ビュー bkst_ordr_shr）のうち、
// モジュールをまたいで使う識別子を表すドメイン層のパッケージ。
package id

// CustomerID は、注文した顧客を特定する識別子（顧客ID。LikeC4 ビュー bkst_ordr_shr_cid）。
type CustomerID string

// String は、顧客ID を文字列として返す。
func (c CustomerID) String() string {
	return string(c)
}

// BookID は、商品カタログコンテキストの書籍を特定する識別子（書籍ID。LikeC4 ビュー bkst_ordr_shr_bid）。
// 値は商品カタログが決め、注文はそのまま使う。
type BookID string

// String は、書籍ID を文字列として返す。
func (b BookID) String() string {
	return string(b)
}

// Package id は、注文コンテキストの共通語彙（LikeC4 ビュー bkst_ordr_shr）のうち、
// モジュールをまたいで使う識別子を表すドメイン層のパッケージ。
package id

import "errors"

// ErrEmpty は、識別子が空であることを表す。
var ErrEmpty = errors.New("識別子が空です")

// CustomerID は、注文した顧客を特定する識別子（顧客ID。LikeC4 ビュー bkst_ordr_shr_cid）。
type CustomerID string

// NewCustomerID は、空でない文字列から顧客ID を作る。
func NewCustomerID(value string) (CustomerID, error) {
	if value == "" {
		return "", ErrEmpty
	}
	return CustomerID(value), nil
}

// String は、顧客ID を文字列として返す。
func (c CustomerID) String() string {
	return string(c)
}

// BookID は、商品カタログコンテキストの書籍を特定する識別子（書籍ID。LikeC4 ビュー bkst_ordr_shr_bid）。
// 値は商品カタログが決め、注文はそのまま使う。
type BookID string

// NewBookID は、空でない文字列から書籍ID を作る。
func NewBookID(value string) (BookID, error) {
	if value == "" {
		return "", ErrEmpty
	}
	return BookID(value), nil
}

// String は、書籍ID を文字列として返す。
func (b BookID) String() string {
	return string(b)
}

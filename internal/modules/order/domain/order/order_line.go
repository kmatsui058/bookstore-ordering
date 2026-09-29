package order

import "github.com/kmatsui058/bookstore-ordering/internal/shared/domain/id"

// OrderLine は、注文の中の1冊の書籍と、その数量と単価（注文明細。LikeC4 ビュー bkst_ordr_ordm_oln）を表す値オブジェクト。
// 数量は1以上、単価は0以上（円）、書籍ID は必須である。
type OrderLine struct {
	bookID    id.BookID
	quantity  int
	unitPrice int
}

// NewOrderLine は、注文明細のルールを満たすときだけ注文明細を作る。
func NewOrderLine(bookID id.BookID, quantity int, unitPrice int) (OrderLine, error) {
	line := OrderLine{bookID: bookID, quantity: quantity, unitPrice: unitPrice}
	if err := line.validate(); err != nil {
		return OrderLine{}, err
	}
	return line, nil
}

// BookID は、注文した書籍の書籍ID を返す。
func (l OrderLine) BookID() id.BookID {
	return l.bookID
}

// Quantity は、数量を返す。
func (l OrderLine) Quantity() int {
	return l.quantity
}

// UnitPrice は、注文したときの単価（円）を返す。
func (l OrderLine) UnitPrice() int {
	return l.unitPrice
}

// Subtotal は、単価 × 数量（円）を返す。
func (l OrderLine) Subtotal() int {
	return l.unitPrice * l.quantity
}

// validate は、注文明細のルールを確かめる。ゼロ値の注文明細が集約に入り込まないよう、集約の側でも呼ぶ。
func (l OrderLine) validate() error {
	if l.bookID == "" {
		return ErrMissingBookID
	}
	if l.quantity < 1 {
		return ErrInvalidQuantity
	}
	if l.unitPrice < 0 {
		return ErrInvalidUnitPrice
	}
	return nil
}

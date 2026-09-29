package api

import (
	"errors"

	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/app/command"
	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/domain/order"
	"github.com/kmatsui058/bookstore-ordering/internal/presentation/api/gen"
)

// エラーのレスポンスの code に入れる、機械向けの値。
const (
	codeInvalidArgument  = "INVALID_ARGUMENT"
	codeNotFound         = "NOT_FOUND"
	codeInvalidState     = "INVALID_STATE"
	codeConcurrentUpdate = "CONCURRENT_UPDATE"
	codeInternal         = "INTERNAL"
)

// isConflict は、注文のいまの状態や同時の変更のために、コマンドを受け付けられなかったエラーかどうかを返す。
func isConflict(err error) bool {
	return errors.Is(err, order.ErrInvalidState) || errors.Is(err, command.ErrConcurrencyConflict)
}

// invalidArgument は、入力が正しくないことを表すエラーのレスポンスを作る。
func invalidArgument(err error) gen.Error {
	return gen.Error{Code: codeInvalidArgument, Message: err.Error()}
}

// notFound は、注文が見つからないことを表すエラーのレスポンスを作る。
func notFound(err error) gen.Error {
	return gen.Error{Code: codeNotFound, Message: err.Error()}
}

// conflict は、状態の衝突を表すエラーのレスポンスを作る。
func conflict(err error) gen.Error {
	if errors.Is(err, command.ErrConcurrencyConflict) {
		return gen.Error{Code: codeConcurrentUpdate, Message: err.Error()}
	}
	return gen.Error{Code: codeInvalidState, Message: err.Error()}
}

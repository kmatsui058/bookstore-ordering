package order

import (
	"errors"
	"fmt"
)

// ErrInvalidArgument は、コマンドの入力が注文のルールを満たさないことを表す。個別のエラーはこれを包む。
var ErrInvalidArgument = errors.New("注文の内容が正しくありません")

// ErrInvalidState は、注文のいまの状態ではそのコマンドを受け付けられないことを表す。個別のエラーはこれを包む。
var ErrInvalidState = errors.New("注文のいまの状態では、この操作はできません")

// ErrNotFound は、指定した注文ID の注文がないことを表す。
var ErrNotFound = errors.New("注文が見つかりません")

// ErrBrokenEventStream は、イベントの列から注文を復元できないことを表す。
var ErrBrokenEventStream = errors.New("注文のイベントの列が壊れています")

var (
	// ErrNoLines は、注文明細が1件もないことを表す。
	ErrNoLines = fmt.Errorf("%w: 注文明細は1件以上必要です", ErrInvalidArgument)
	// ErrInvalidQuantity は、注文明細の数量が1未満であることを表す。
	ErrInvalidQuantity = fmt.Errorf("%w: 数量は1以上です", ErrInvalidArgument)
	// ErrInvalidUnitPrice は、注文明細の単価が負であることを表す。
	ErrInvalidUnitPrice = fmt.Errorf("%w: 単価は0以上です", ErrInvalidArgument)
	// ErrMissingBookID は、注文明細に書籍ID がないことを表す。
	ErrMissingBookID = fmt.Errorf("%w: 書籍ID は必須です", ErrInvalidArgument)
	// ErrMissingCustomerID は、顧客ID がないことを表す。
	ErrMissingCustomerID = fmt.Errorf("%w: 顧客ID は必須です", ErrInvalidArgument)
	// ErrInvalidPaymentMethod は、支払方法が定義された値ではないことを表す。
	ErrInvalidPaymentMethod = fmt.Errorf("%w: 支払方法が正しくありません", ErrInvalidArgument)

	// ErrNotConfirmable は、受付済みではない注文を確定しようとしたことを表す。
	ErrNotConfirmable = fmt.Errorf("%w: 確定できるのは受付済みの注文だけです", ErrInvalidState)
	// ErrNotCancellable は、出荷前ではない注文をキャンセルしようとしたことを表す。
	ErrNotCancellable = fmt.Errorf("%w: キャンセルできるのは出荷前（受付済み・確定）の注文だけです", ErrInvalidState)
	// ErrNotShippable は、確定ではない注文を出荷済みにしようとしたことを表す。
	ErrNotShippable = fmt.Errorf("%w: 出荷済みにできるのは確定した注文だけです", ErrInvalidState)
)

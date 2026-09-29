package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/app/command"
	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/app/query"
	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/domain/order"
	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/infra/messaging/inprocess"
	"github.com/kmatsui058/bookstore-ordering/internal/modules/order/infra/repository/inmemory"
	"github.com/kmatsui058/bookstore-ordering/internal/presentation/api/gen"
)

// NewHTTPHandler は、依存を組み立てて（DI）、注文 API の HTTP ハンドラーを作る。
// 永続化はメモリのストア、イベントの公開はプロセス内の公開者で、公開したイベントはログに出す。
func NewHTTPHandler(logger *slog.Logger) http.Handler {
	store := inmemory.NewStore()
	publisher := inprocess.NewPublisher(logEvent(logger))
	commands := command.NewService(inmemory.NewTransactionManager(store), publisher)
	queries := query.NewService(store)
	strict := gen.NewStrictHandlerWithOptions(NewHandler(commands, queries), nil, gen.StrictHTTPServerOptions{
		RequestErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, err error) {
			writeError(w, http.StatusBadRequest, invalidArgument(err))
		},
		ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			logger.ErrorContext(r.Context(), "リクエストの処理に失敗しました", "method", r.Method, "path", r.URL.Path, "error", err)
			writeError(w, http.StatusInternalServerError, gen.Error{Code: codeInternal, Message: "サーバーで問題が起きました"})
		},
	})
	return gen.HandlerWithOptions(strict, gen.StdHTTPServerOptions{
		ErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, err error) {
			writeError(w, http.StatusBadRequest, invalidArgument(err))
		},
	})
}

// logEvent は、公開されたドメインイベントをログに出す購読者を作る。
func logEvent(logger *slog.Logger) inprocess.Subscriber {
	return func(ctx context.Context, e order.Event) error {
		logger.InfoContext(ctx, "ドメインイベントを公開しました",
			"eventType", e.EventType(),
			"orderID", e.OrderID().String(),
			"version", e.Metadata().Version(),
			"correlationID", e.Metadata().CorrelationID().String(),
		)
		return nil
	}
}

// writeError は、エラーのレスポンスを JSON で書く。
func writeError(w http.ResponseWriter, status int, body gen.Error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

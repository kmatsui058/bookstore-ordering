// Package order は、注文コンテキストの注文管理モジュール（LikeC4 ビュー bkst_ordr_ordm）のうち、
// 集約ルート「注文」（LikeC4 ビュー bkst_ordr_ordm_ord）を表すドメイン層のパッケージ。
// 注文の不変条件と状態の遷移（LikeC4 ビュー bkst_ordr_ordm_ord_state）を守り、状態を変えるたびにドメインイベントを発行する。
// 永続化や外部への伝送は行わない。
package order

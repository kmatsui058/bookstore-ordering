// Package gen は、契約リポジトリの注文 API の OpenAPI の定義（contracts/ordering/definitions/openapi/openapi.yaml）から
// oapi-codegen で生成した型とサーバーのインターフェースを置くプレゼンテーション層のパッケージ。
// server.gen.go は生成物なので手で直さない。定義を変えるときは契約リポジトリで変え、生成し直す。
package gen

//go:generate go tool oapi-codegen -config ../../../../configs/oapi-codegen.yaml ../../../../contracts/ordering/definitions/openapi/openapi.yaml

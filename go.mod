module github.com/akaritrading/platform

go 1.14

require (
	github.com/akaritrading/backtest/pkg v0.0.0
	github.com/akaritrading/engine/pkg v0.0.0
	github.com/akaritrading/libs v0.0.0
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/gorilla/websocket v1.4.2
	github.com/jinzhu/gorm v1.9.16
	github.com/lib/pq v1.8.0 // indirect
)

replace github.com/akaritrading/libs v0.0.0 => ../libs
replace github.com/akaritrading/backtest/pkg v0.0.0 => ../backtest/pkg
replace github.com/akaritrading/engine/pkg v0.0.0 => ../engine/pkg


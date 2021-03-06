module github.com/akaritrading/platform

go 1.15

require (
	github.com/akaritrading/backtest v0.0.3
	github.com/akaritrading/libs v0.0.6
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/lib/pq v1.8.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sendgrid/rest v2.6.1+incompatible // indirect
	github.com/sendgrid/sendgrid-go v3.6.3+incompatible
	gorm.io/gorm v1.20.2 // indirect
)

replace github.com/akaritrading/backtest v0.0.0 => ../backtest

replace github.com/akaritrading/libs v0.0.0 => ../libs

// replace github.com/akaritrading/prices v0.0.0 => ../prices
// replace github.com/akaritrading/engine v0.0.0 => ../engine

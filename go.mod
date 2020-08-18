module github.com/akaritrading/platform

go 1.14

require (
	github.com/akaritrading/libs v0.0.0
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/jinzhu/gorm v1.9.16
	github.com/lib/pq v1.8.0 // indirect
)

replace github.com/akaritrading/libs v0.0.0 => ../libs

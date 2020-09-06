module github.com/akaritrading/platform

go 1.14

require (
	github.com/akaritrading/libs v0.0.0
	github.com/akaritrading/platform/pkg/engine v0.0.0
	github.com/go-chi/chi v4.1.2+incompatible
	github.com/jinzhu/gorm v1.9.16
	github.com/lib/pq v1.8.0 // indirect
	github.com/satori/go.uuid v1.2.0
	github.com/sendgrid/rest v2.6.1+incompatible // indirect
	github.com/sendgrid/sendgrid-go v3.6.3+incompatible
	golang.org/x/crypto v0.0.0-20200728195943-123391ffb6de
	gorm.io/gorm v1.9.19 // indirect
)

replace github.com/akaritrading/libs v0.0.0 => ../libs

replace github.com/akaritrading/platform/pkg/engine v0.0.0 => ./pkg

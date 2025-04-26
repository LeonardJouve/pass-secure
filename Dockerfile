FROM golang:1.24

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOOS=linux go build -o /app/pass-secure main.go

CMD ["/app/pass-secure"]

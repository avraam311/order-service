FROM golang:alpine AS build_base

WORKDIR /app

COPY ./go.mod ./go.sum ./

RUN go mod download

COPY . .

RUN go build -o app ./backend/cmd/app/main.go

FROM alpine AS runner

COPY --from=build_base /app/app .
COPY ./.env .
COPY ./backend/config/config.yaml ./config/config.yaml

CMD ["./app"]
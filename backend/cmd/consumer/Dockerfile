FROM golang:alpine AS build_base

WORKDIR /app

COPY ./go.mod ./go.sum ./

RUN go mod download

COPY . .

RUN go build -o consumer ./backend/cmd/consumer/main.go

FROM alpine AS runner

COPY --from=build_base /app/consumer .
COPY ./.env .
COPY ./backend/config/config.yaml ./config/config.yaml

CMD ["./consumer"]
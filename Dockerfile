FROM golang:1.22.1-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN go build -ldflags="-s -w" -o /itpgb ./cmd/itpg-backend/main.go

FROM scratch

WORKDIR /

COPY --from=build /itpgb /itpgb

EXPOSE 5555

ENTRYPOINT ["/itpgb"]

CMD ["--help"]

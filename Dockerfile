FROM golang:1.22.1-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN go build -ldflags="-s -w" -o /itpg ./cmd/itpg-backend/main.go

FROM scratch

WORKDIR /

COPY --from=build /itpg /itpg

EXPOSE 5555

ENTRYPOINT ["/itpg"]

CMD ["--help"]

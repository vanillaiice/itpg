FROM golang:1.23.1-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN go build -ldflags="-s -w" -o /itpg .

FROM scratch

LABEL org.opencontainers.image.source=https://github.com/vanillaiice/itpg

WORKDIR /

COPY --from=build /itpg /itpg

EXPOSE 5555

ENTRYPOINT ["/itpg"]

CMD ["--help"]

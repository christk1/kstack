FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/demo ./main.go

FROM alpine:3
COPY --from=build /out/demo /usr/local/bin/demo
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/demo"]

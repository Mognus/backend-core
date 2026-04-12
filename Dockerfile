FROM golang:1.26-alpine AS builder

COPY services/ /services/

WORKDIR /backend

COPY backend-core/go.mod backend-core/go.sum* ./
RUN go mod download

COPY backend-core/ .

RUN go build -ldflags="-s -w" -o /out/server ./cmd/server


FROM alpine:3.21

WORKDIR /app

COPY --from=builder /out/server ./server

EXPOSE 8080

CMD ["./server"]

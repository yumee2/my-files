FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /out/file-uploader .

FROM alpine:3.22

WORKDIR /app

COPY --from=build /out/file-uploader /app/file-uploader

EXPOSE 8000

CMD ["./file-uploader"]

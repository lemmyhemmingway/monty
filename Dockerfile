FROM golang:1.23-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o monty

FROM alpine:3.18
WORKDIR /app
COPY --from=build /app/monty .
COPY --from=build /app/templates ./templates
EXPOSE 3000
CMD ["./monty"]

FROM node:20-alpine AS frontend-build
WORKDIR /app
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.23-alpine AS backend-build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Copy built React files to static
COPY --from=frontend-build /app/dist/* ./static/
RUN CGO_ENABLED=0 GOOS=linux go build -o monty

FROM alpine:3.18
WORKDIR /app
COPY --from=backend-build /app/monty .
COPY --from=backend-build /app/templates ./templates
COPY --from=backend-build /app/static ./static
EXPOSE 3000
CMD ["./monty"]

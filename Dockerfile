# syntax=docker/dockerfile:1

FROM node:22-alpine AS frontend
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
WORKDIR /src
COPY frontend-admin/package.json ./
RUN npm install --registry=https://registry.npmmirror.com
COPY frontend-admin/ ./
RUN npm run build

FROM golang:1.23-alpine AS backend
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
ENV GOPROXY=https://goproxy.cn,direct \
    CGO_ENABLED=0 \
    GOOS=linux
WORKDIR /src
COPY backend/go.mod ./
RUN go mod download
COPY backend/ ./
RUN go build -ldflags="-s -w" -o /broker ./cmd/broker

FROM alpine:3.20
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache wget ca-certificates tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo Asia/Shanghai > /etc/timezone \
    && adduser -D -u 1001 kafka \
    && mkdir -p /data /app/web \
    && chown -R kafka:kafka /data /app
WORKDIR /app
COPY --from=backend /broker /usr/local/bin/broker
COPY --from=frontend /src/dist /app/web
ENV TZ=Asia/Shanghai \
    STATIC_DIR=/app/web \
    DATA_DIR=/data \
    HTTP_ADDR=:8080
EXPOSE 8080
USER kafka
CMD ["broker"]

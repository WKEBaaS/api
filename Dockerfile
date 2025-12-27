# === Build Stage ===
FROM golang:1.25.4-alpine AS builder

WORKDIR /app

# 下載模組依賴 (利用 Docker Layer Caching)
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 優化編譯指令：
# -ldflags="-s -w" : 移除除錯資訊與符號表，大幅縮小檔案體積
# -o /app/server   : 統一輸出檔名
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server .

# === Production Stage ===
# 使用純 Alpine，不包含 Go SDK，體積極小 (~5MB)
FROM alpine:3.23 AS production

# 安裝基礎憑證與時區資料 (若程式需要發送 HTTPS 請求或處理時間)
RUN apk --no-cache add ca-certificates=20251003-r0

WORKDIR /app

# 建立非 root 使用者以提升安全性
RUN addgroup -S app && adduser -S app -G app

# 從 Builder 階段只複製編譯好的執行檔
COPY --from=builder /app/server /app/server

# 切換到非 root 使用者
USER app

EXPOSE 8080

# 執行程式 (確保檔名與 build 階段一致)
CMD ["/app/server"]

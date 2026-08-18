FROM --platform=$BUILDPLATFORM golang:1.22-bookworm AS backend
ARG TARGETOS TARGETARCH
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GOTOOLCHAIN=local go build -trimpath -o /out/windops ./cmd/server

FROM node:22-bookworm-slim AS frontend
WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=backend /out/windops /app/windops
COPY --from=frontend /src/web/dist /app/web
EXPOSE 8080
ENV WINDOPS_ADDR=:8080 WINDOPS_DB=/data/windops.db WINDOPS_WEB=/app/web
VOLUME ["/data"]
ENTRYPOINT ["/app/windops"]


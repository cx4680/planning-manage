FROM image.cestc.cn/secdb/golang:1.20-cert-mgr as builder
ENV CGO_ENABLED=0

WORKDIR /build
COPY . .
RUN cd cmd && go build -o ../planning-manage

FROM image.cestc.cn/baseos/base/cclinux2209:22.09.2-3
WORKDIR /app

COPY --from=builder /build/planning-manage ./planning-manage
COPY --from=builder /build/migrations/*.sql ./migrations/

ENTRYPOINT ["./planning-manage"]

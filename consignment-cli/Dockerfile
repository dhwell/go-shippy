FROM golang:1.13.8 as builder

ENV GOPROXY https://goproxy.cn,direct
ENV GO111MODULE on
ENV GOSUMDB off

#下载依赖包
WORKDIR /go/cache
ADD go.mod .
ADD go.sum .
RUN go mod download

WORKDIR /go/release

ADD . .

#编译应用
RUN GOOS=linux CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix cgo -o app main.go

FROM scratch

COPY --from=builder /go/release/app /

CMD ["./app"]

FROM golang:1.10.1 AS go-builder
WORKDIR /go
RUN go get github.com/wcharczuk/go-chart
COPY picasso.go /go
RUN CGO_ENABLE=0 GOOS=linux GOARCH=amd64 go build -o picasso .

FROM loadimpact/k6
WORKDIR /kubefaas-bench
COPY --from=go-builder /go/picasso /usr/local/bin/picasso
RUN apk --update add --no-cache bash curl
RUN curl -Lo kubefaas https://github.com/srcmesh/kubefaas/releases/download/$(curl --silent "https://api.github.com/repos/srcmesh/kubefaas/releases/latest" | grep "tag_name" |sed -E 's/.*"([^"]+)".*/\1/')/kubefaas-cli-linux && chmod +x kubefaas && mv kubefaas /usr/local/bin/

ENTRYPOINT ["sh"]

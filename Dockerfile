FROM golang:1.8
WORKDIR /go/src/github.com/lawrencegripper/kube-azureresources
COPY . .
# RUN curl https://glide.sh/get | sh
# RUN glide install -v
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o controller .

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /go/src/github.com/lawrencegripper/kube-azureresources/controller .
CMD ["./controller"]  

FROM golang:1.10.3 as builder

WORKDIR /go-go-docker
COPY . .
RUN go get -d -v "github.com/docker/docker/api/types"
RUN go get -d -v "github.com/docker/docker/client"
RUN go get -d -v "github.com/gin-contrib/static"
RUN go get -d -v "github.com/gin-gonic/gin"
RUN go get -d -v "golang.org/x/net/context"
RUN go get -d -v "gopkg.in/olahol/melody.v1"
RUN CGO_ENABLED=0 GOOS=linux go build -o go-go-docker .

FROM docker:18.05.0-ce-dind
EXPOSE 3000
COPY --from=builder /go-go-docker .
CMD ["./go-go-docker"]
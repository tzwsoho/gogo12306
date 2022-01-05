FROM golang as build

ENV CGOENABLED=0
ENV GO111MODULE=on
ENV GOPROXY=http://goproxy.io,direct

WORKDIR /gogo12306
COPY . .

RUN go build -tags netgo

FROM apline

WORKDIR /gogo12306

COPY --from=build /gogo12306/config.json /gogo12306/config.json
COPY --from=build /gogo12306/gogo12306 /gogo12306/gogo12306

CMD ["/gogo12306/gogo12306", "-g"]
FROM golang as build

ENV CGOENABLED=0
ENV GO111MODULE=on
ENV GOPROXY=http://goproxy.io,direct

WORKDIR /gogo12306
COPY . .

RUN go build -tags netgo

FROM apline

COPY --from=build /gogo12306/bin/gogo12306 /gogo12306/gogo12306
COPY --from=build /gogo12306/bin/cdnfilter /gogo12306/cdnfilter
COPY --from=build /gogo12306/config/cdn.txt /gogo12306/cdn.txt
COPY --from=build /gogo12306/config/config.json /gogo12306/config.json

CMD ["/gogo12306/gogo12306", "-c", "-r"]
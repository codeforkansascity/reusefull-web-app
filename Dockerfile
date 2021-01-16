FROM alpine as build
# install build tools
RUN apk add go git
# cache dependencies
ADD go.mod go.sum ./
RUN go env -w GOPROXY=direct
# RUN go mod download GOPROXY=direct
# build
ADD . .
RUN go build -o /main
# copy artifacts to a clean image
FROM alpine
COPY --from=build /main /main
ENTRYPOINT [ "/main" ]

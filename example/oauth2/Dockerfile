FROM golang:alpine AS builder
WORKDIR /brock/example/
COPY . ./
RUN apk --no-cache add ca-certificates git
# PAT with scopes [read_repository]
RUN git config --global url."https://glpat-3mk8zATKbBhzu2x-oFpD@gitlab.com".insteadOf "https://gitlab.com"
RUN go env -w GONOPROXY=gitlab.com/brick-io
RUN go env -w GONOSUMDB=gitlab.com/brick-io
RUN go env -w GOPRIVATE=gitlab.com/brick-io
RUN go build -o oauth2 .

FROM alpine
WORKDIR /root/
COPY --from=builder /brock/example/oauth2 ./
CMD ["./oauth2"]

# Uses a multi-stage build to decrease the size of final image.
#
# The first stage installs packages and libraries which are required for the
# build-process only. The second stage references the first stage (see 'as
# builder') to copy the built executable to the final stage.


#
# build-bin stage
#

# Always use the full path as base image.
FROM docker.io/golang:alpine as builder
# Add gcc and musl-dev so go-libs which consist of c-code (e.g. sqlite3) can be
# compiled.
RUN apk add gcc musl-dev
# Copy the current directory (".") to the image under the path "/build".
COPY . /build
# Change to the /build directory and compile the go project. The binary can
# be found at /build/video-platform-from-scratch
RUN cd /build && go build -o video-platform-from-scratch cmd/server.go

#
# build-img stage
#

FROM docker.io/alpine
# Update any packages and add the default certificate bundle.
RUN apk update && apk add ca-certificates
# Copy the go-binary from the build-stage and put it into a bin directory which
# is part of our PATH variable.
COPY --from=builder /build/video-platform-from-scratch /usr/local/bin
# Copy the views and assets into the container.
ADD views /data/views
ADD assets /data/assets
# Create database storage directory.
RUN mkdir -p /data/db
# Let the /data directory be the default work directory.
WORKDIR /data
# Set the go binary to be the entrypoint, so it does not need to be specified
# explicitly when starting the container.
ENTRYPOINT ["video-platform-from-scratch"]


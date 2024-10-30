## build ypsdb binary
FROM docker.io/golang:1.23-alpine AS build-env

RUN apk upgrade -U --force-refresh --no-cache && apk add --no-cache --purge --clean-protected -l -u make git

# copy ypsdb source
WORKDIR /go/src/github.com/YPS-Database/yps-db-backend
COPY . .

# compile
RUN make install

## build ypsdb container
FROM docker.io/alpine:3.19

# metadata
LABEL maintainer="Daniel Oaks <daniel@danieloaks.net>" \
      description="The YPS Database is a database and resource library of publicly available materials relating to the Youth, Peace and Security agenda"

# standard port listened on
EXPOSE 8465/tcp

# ypsdb itself
COPY --from=build-env /go/bin/yps-db-backend \
                      /go/src/github.com/YPS-Database/yps-db-backend/distrib/docker/run.sh \
                      /ypsdb-bin/
COPY --from=build-env /go/src/github.com/YPS-Database/yps-db-backend/migrations \
                      /ypsdb-bin/migrations

# launch
ENTRYPOINT ["/ypsdb-bin/run.sh"]

# # uncomment to debug
# RUN apk add --no-cache bash
# RUN apk add --no-cache vim
# CMD /bin/bash

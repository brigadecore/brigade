FROM node:12.16.2-alpine3.11
ARG VERSION
ENV WORKER_VERSION=${VERSION}

RUN apk add bash && npm install -g typescript

WORKDIR /var/brigade-worker/src/brigadier
COPY v2/brigadier/ .

# Can't figure out how to do this in POSIX, so bash it is...
RUN bash -c 'if [[ $WORKER_VERSION =~ ^v[0-9]+(\.[0-9]+)*(\-.+)?$ ]]; then \
    V=$(echo $WORKER_VERSION | cut -c 2-); \
  else \
    V=0.0.1 ; \
  fi ; \
  sed -i s/placeholder/$V/ package.json && yarn install && yarn build \
  '

WORKDIR /var/brigade-worker/src/worker
COPY v2/worker/ .
RUN yarn install && yarn build

CMD yarn -s start

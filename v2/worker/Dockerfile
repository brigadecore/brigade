FROM node:16.11.0-alpine3.14
ARG VERSION
ENV WORKER_VERSION=${VERSION}

RUN apk add bash

# Prevent update notices from appearing on npm/yarn install/run, etc.
RUN npm config set update-notifier false \
  && yarn config set disable-self-update-check true

WORKDIR /var/brigade-worker/brigadier
COPY v2/brigadier/ .
RUN bash -c 'if [[ $WORKER_VERSION =~ ^v[0-9]+(\.[0-9]+)*(\-.+)?$ ]]; then \
      sed -i s/0.0.1-placeholder/$(echo $WORKER_VERSION | cut -c 2-)/ package.json ; \
    fi ; \
    yarn install --prod && yarn build \
  '

WORKDIR /var/brigade-worker/brigadier-polyfill
COPY v2/brigadier-polyfill/ .
RUN bash -c 'if [[ $WORKER_VERSION =~ ^v[0-9]+(\.[0-9]+)*(\-.+)?$ ]]; then \
      sed -i s/0.0.1-placeholder/$(echo $WORKER_VERSION | cut -c 2-)/ package.json ; \
    fi ; \
    yarn install --prod && yarn build \
  '

WORKDIR /var/brigade-worker/worker
COPY v2/worker/ .
RUN bash -c 'if [[ $WORKER_VERSION =~ ^v[0-9]+(\.[0-9]+)*(\-.+)?$ ]]; then \
      sed -i s/0.0.1-placeholder/$(echo $WORKER_VERSION | cut -c 2-)/ package.json ; \
    fi ; \
    yarn install --prod && yarn build \
  '

CMD yarn -s start

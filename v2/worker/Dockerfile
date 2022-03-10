FROM --platform=$BUILDPLATFORM debian:bullseye-slim

ARG VERSION

WORKDIR /var/brigade-worker/brigadier
COPY v2/brigadier/ .
RUN bash -c 'if [[ $VERSION =~ ^v[0-9]+(\.[0-9]+)*(\-.+)?$ ]]; then \
      sed -i s/0.0.1-placeholder/$(echo $VERSION | cut -c 2-)/ package.json ; \
    fi \
  '

WORKDIR /var/brigade-worker/brigadier-polyfill
COPY v2/brigadier-polyfill/ .
RUN bash -c 'if [[ $VERSION =~ ^v[0-9]+(\.[0-9]+)*(\-.+)?$ ]]; then \
      sed -i s/0.0.1-placeholder/$(echo $VERSION | cut -c 2-)/ package.json ; \
    fi \
  '

WORKDIR /var/brigade-worker/worker
COPY v2/worker/ .
RUN bash -c 'if [[ $VERSION =~ ^v[0-9]+(\.[0-9]+)*(\-.+)?$ ]]; then \
      sed -i s/0.0.1-placeholder/$(echo $VERSION | cut -c 2-)/ package.json ; \
    fi \
  '

FROM $TARGETARCH/node:16.14.0-alpine3.15

# Prevent update notices from appearing on npm/yarn install/run, etc.
RUN npm config set update-notifier false \
  && yarn config set disable-self-update-check true

WORKDIR /var/brigade-worker/brigadier
COPY --from=0 /var/brigade-worker/brigadier/ .
RUN yarn install --prod && yarn build

WORKDIR /var/brigade-worker/brigadier-polyfill
COPY --from=0 /var/brigade-worker/brigadier-polyfill/ .
RUN yarn install --prod && yarn build

WORKDIR /var/brigade-worker/worker
COPY --from=0 /var/brigade-worker/worker/ .
RUN yarn install --prod && yarn build

CMD yarn -s start

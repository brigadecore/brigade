FROM --platform=$BUILDPLATFORM node:16.11.0-bullseye-slim as builder

ARG VERSION

WORKDIR /var/brigade-worker/brigadier
COPY v2/brigadier/ .
RUN bash -c 'if [[ $VERSION =~ ^v[0-9]+(\.[0-9]+)*(\-.+)?$ ]]; then \
      sed -i s/0.0.1-placeholder/$(echo $VERSION | cut -c 2-)/ package.json ; \
    fi ; \
    yarn install --prod && yarn build \
  '

WORKDIR /var/brigade-worker/brigadier-polyfill
COPY v2/brigadier-polyfill/ .
RUN bash -c 'if [[ $VERSION =~ ^v[0-9]+(\.[0-9]+)*(\-.+)?$ ]]; then \
      sed -i s/0.0.1-placeholder/$(echo $VERSION | cut -c 2-)/ package.json ; \
    fi ; \
    yarn install --prod && yarn build \
  '

WORKDIR /var/brigade-worker/worker
COPY v2/worker/ .
RUN bash -c 'if [[ $VERSION =~ ^v[0-9]+(\.[0-9]+)*(\-.+)?$ ]]; then \
      sed -i s/0.0.1-placeholder/$(echo $VERSION | cut -c 2-)/ package.json ; \
    fi ; \
    yarn install --prod && yarn build \
  '

# Prevent update notices from appearing on npm/yarn install/run, etc.
# We can't run this in the final stage because its base image lacks a shell, so
# we run it here and copy the resulting files over in the final stage.
RUN npm config set update-notifier false && \
  yarn config set disable-self-update-check true

FROM gcr.io/distroless/nodejs:16 as final

ARG VERSION
ENV WORKER_VERSION=${VERSION}

# Prevent update notices from appearing on npm/yarn install/run, etc.
COPY --from=builder --chown=nonroot:nonroot /root/.npmrc /home/nonroot/
COPY --from=builder --chown=nonroot:nonroot /usr/local/share/.yarnrc /home/nonroot/

COPY --from=builder --chown=nonroot:nonroot /var/brigade-worker/brigadier/ /var/brigade-worker/brigadier/

COPY --from=builder --chown=nonroot:nonroot /var/brigade-worker/brigadier-polyfill/ /var/brigade-worker/brigadier-polyfill/

WORKDIR /var/brigade-worker/worker
COPY --from=builder --chown=nonroot:nonroot /var/brigade-worker/worker/ .

CMD ["/var/brigade-worker/worker/dist/prestart.js"]

USER nonroot

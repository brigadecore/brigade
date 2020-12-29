FROM node:12.16.2-alpine3.11
ARG VERSION
ENV WORKER_VERSION=${VERSION}
WORKDIR /src
COPY v2/brigadier/ brigadier/
COPY v2/worker/ worker/
WORKDIR /src/brigadier
RUN yarn install
WORKDIR /src/worker
RUN yarn install && yarn build

CMD yarn -s start

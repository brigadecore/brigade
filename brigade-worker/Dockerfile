FROM {{.docker}}/node:8-alpine
{{.qemu}}

WORKDIR /home/src
COPY . /home/src
RUN yarn install && yarn build && yarn cache clean

CMD yarn run test

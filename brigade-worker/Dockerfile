FROM ${DOCKER_ARCH}/node:8-alpine
${QEMU_IMPORT}

WORKDIR /home/src
COPY . /home/src
RUN yarn install && yarn build

CMD yarn run test

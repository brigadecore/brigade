FROM node:12.22.7-alpine3.14

WORKDIR /home/src
COPY brigade-worker/ /home/src/
RUN yarn install && yarn build

CMD yarn run test

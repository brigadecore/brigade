FROM node:8-alpine

WORKDIR /home/src
COPY brigade-worker/ /home/src/
RUN yarn install && yarn build

CMD yarn run test

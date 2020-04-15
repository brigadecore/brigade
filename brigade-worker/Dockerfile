FROM node:10-alpine

WORKDIR /home/src
COPY brigade-worker/ /home/src/
RUN yarn build

CMD yarn run test

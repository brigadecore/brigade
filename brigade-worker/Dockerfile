FROM node:12.16.2-alpine3.11

WORKDIR /home/src
COPY brigade-worker/ /home/src/
RUN yarn build

CMD yarn run test

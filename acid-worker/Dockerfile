FROM node:8

WORKDIR /home/src
COPY . /home/src
RUN yarn install && yarn build

CMD yarn run test

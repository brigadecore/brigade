// Inspired by https://github.com/radu-matei/brigade-javascript-deps

const { events } = require("@brigadecore/brigadier");

// This dependency is declared in the package.json file
const is = require("is-thirteen");

// This dependency exists relative to the project's git repository
const circle = require("./circle");

events.on("brigade.sh/cli", "exec", async event => {
  console.log("is 13 thirteen? " + is(13).thirteen());
  console.log("area of a circle with radius 3: " + circle.area(3));
});

events.process();

const { uniqueNamesGenerator, adjectives, animals } = require('unique-names-generator');

const randomName = uniqueNamesGenerator({
  dictionaries: [adjectives, animals],
  length: 2,
  separator: "-"
});

console.log(`Hello, ${randomName}!`);

const { events } = require("brigadier");
const XML = require("xml-simple");
const { alpineJob } = require("./mylib");

events.on("exec", () => {
  XML.parse("<say><to>world</to></say>", (e, say) => {
    console.log(`Saying hello to ${say.to}`);
  })

  const alpine = alpineJob("myjob");
  alpine.run();
});

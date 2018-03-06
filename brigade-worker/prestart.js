const process = require("process")
const fs = require("fs")
const path = require("path")

// Worker should always set both env vars. The defaults are for local testing.
const script = process.env.BRIGADE_SCRIPT || "brigade.js"
const scriptDir = process.env.BRIGADE_SCRIPT_DIR || "/etc/brigade-scripts/"
const outputDir = process.env.BRIGADE_OUTPUT_DIR || "dist/"
const vcsScript = process.env.BRIGADE_VCS_SCRIPT || "/vcs/brigade.js"

try {
  fs.readdir(scriptDir, (err, files) => {
    files.forEach(file => {
      if (file == script) {
        var data = loadScript(path.join(scriptDir, script))
        let wrapper = "const {overridingRequire} = require('./require');((require) => {" +
          data.toString() +
          "})(overridingRequire)";
        fs.writeFile(path.join(outputDir, script), wrapper, () => {
          console.log("prestart: dist/brigade.js written");
        });
      } else {
        copyFile(path.join(scriptDir, file), path.join(outputDir, file), function(err) {
          if (err != null) {
            console.log("prestart:", err);
          }
        });
      }
    });
  });
} catch (e) {
  console.log("prestart: no script override")
  process.exit(1)
}

function copyFile(source, target, cb) {
  var cbCalled = false;
  var rd = fs.createReadStream(source);
  rd.on("error", function(err) {
    done(err);
  });

  var wr = fs.createWriteStream(target);
  wr.on("error", function(err) {
    done(err);
  });
  wr.on("close", function(ex) {
    done();
  });

  rd.pipe(wr);

  function done(err) {
    if (!cbCalled) {
      cb(err);
      cbCalled = true;
    }
  }
}

// loadScript tries to load the configured script. But if it can't, it falls
// back to the VCS copy of the script.
function loadScript(script) {
  // This happens if the secret volume is mis-mounted, which should never happen.
  if (!fs.existsSync(script)) {
    console.log("prestart: no script found. Falling back to VCS script")
    return fs.readFileSync(vcsScript, 'utf8')
  }
  var data = fs.readFileSync(script, 'utf8')
  if (data == "") {
    // This happens if no file was submitted by the consumer.
    console.log("prestart: empty script found. Falling back to VCS script")
    return fs.readFileSync(vcsScript, 'utf8')
  }
  return data
}

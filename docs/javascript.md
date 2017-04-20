# Acid JavaScript

Acid JavaScript is a dialect of JavaScript for writing Acid build files.

Acid JavaScript has access to a few libraries:

- The Underscore.js library is built into Acid.js
- AcidJS has a number of built-in objects.

### The Job object

To create a new job:

```javascript
j = new Job(name)
```

Parameters:

- A job name (alpha-numeric characters plus dashes).

Properties:

- `image`: A Docker image with optional tag.
- `env`: Key/value pairs that will be injected into the environment. The key is
  the variable name (`MY_VAR`), and the value is the string value (`foo`)
- `secrets`: Key/value pairs where the key is the name of the environment variable
  and the value is the name of the item in the Secret. `{ "DB_PASS": "dbpassword" }`

Methods:

- `run(pushRecord)`: Run this job.

## Acid JS and ECMAScript (and browser-based JS)

Acid JS is ECMAScript 5. It has a few differences, though.

- The Regular Expression library is Go's regular expression library
- It does not provide `setTimeout` or `setInterval`
- Browser objects, like `window`, are not present

# Brigade Worker: The Brigade script runner

Brigade Worker is part of the Brigade system. It is responsible for executing an
`brigade.js` task.

The Brigade worker accepts an `brigade.js` script along with event and project information
and it executes the Brigade script.

The `brigade.js` script has access to all of the exported fields in `brigadier`.

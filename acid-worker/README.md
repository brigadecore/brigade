# Acid Worker: The Acid script runner

Acid Worker is part of the Acid system. It is responsible for executing an
`acid.js` task.

The Acid worker accepts an `acid.js` script along with event and project information
and it executes the Acid script.

The `acid.js` script has access to all of the exported fields in `libacid`.

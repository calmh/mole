"use strict";

var spawn = require('child_process').spawn;

module.exports = function passthroughSpawn(cmd, args) {
    if (process.version.match(/^v0\.6\./)) {
        // Earliest node version supported. Requires use of deprecated customFds options.
        return spawn(cmd, args, { customFds: [ 0, 1, 2 ] });
    } else {
        // Modern node. Uses stdio: inherit.
        return spawn(cmd, args, { stdio: 'inherit' });
    }
}


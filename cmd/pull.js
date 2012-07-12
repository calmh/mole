"use strict";

var fs = require('fs');
var path = require('path');
var version = require('version');

var con = require('../lib/console');
var tun = require('../lib/tunnel');

module.exports = pull;
pull.help = 'Get tunnel definitions from the server';
pull.prio = 1;

function pull(opts, state) {
    // Update the local file repository with newer data from the server.

    updateFromServer(opts, state);

    // The user presumably want's to be up to date so check that we're running
    // the latest version.

    checkVersion(opts, state);
}

function updateFromServer(opts, state) {

    // Get the list of tunnel definitions from the server. The list includes
    // (name, mtime) for each tunnel. We'll use the `mtime` to figure out if we
    // need to download the definition or not.

    con.debug('Requesting tunnel list from server');
    state.client.list(function (result) {
        con.debug('Got ' + result.length + ' entries');

        // We use this to keep track of the number of outstanding requests and
        // to print a message when every request has finished.

        var inProgress = 0;
        function done() {
            inProgress -= 1;
            if (inProgress === 0) {
                con.ok(result.length + ' tunnel definitions in sync');
                inProgress = -1;
            }
        }

        result.forEach(function (res) {
            inProgress += 1;

            // Figure out the local filename that corresponds to this tunnel, if we have it.

            var local = path.join(state.path.tunnels, res.name);

            // We need to fetch the file only if we either don't already have
            // it, or if the mtime as sent by the server differs from what we
            // have locally.

            var fetch = false;
            if (!fs.existsSync(local)) {
                fetch = true;
            } else {
                var s = fs.statSync(local);
                if (s.mtime.getTime() !== parseInt(res.mtime, 10)) {
                    fetch = true;
                }
            }

            // If we need to fetch a tunnel definition, send a server request to do so.

            if (fetch) {
                state.client.saveBin(res.name, local, function () {

                    // When the request completes, we set the mtime to match
                    // that sent by the server. The server sends it in
                    // milliseconds since that's what Javascript timestamps
                    // usually are, but utimesSync expects seconds.

                    var mtime = Math.floor(res.mtime / 1000);
                    fs.utimesSync(local, mtime, mtime);

                    con.info('Pulled ' + tun.name(res.name));

                    // Mark this request as completed, print out status if it was the last one.

                    done();
                });
            } else {

                // Mark this request as completed. We don't do it immediately
                // since that would result in the `inProgress` counter flapping
                // between 1 and 0 when we didn't need to fetch anything.
                // Instead, queue the call for the next tick, when `inProgress`
                // has been incremented all the way.

                process.nextTick(done);
            }
        });
    });
}

// Fetch the latest version number for mole from the npm repository and print a
// 'time to upgrade'-message if there's a mismatch.

function checkVersion(opts, state) {
    version.fetch('mole', function (err, ver) {
        if (!err && ver) {
            if (ver !== state.pkg.version) {
                con.info('You are using mole v' + state.pkg.version + '; the latest version is v' + ver);
                con.info('Use "sudo npm -g update mole" to upgrade');
            } else {
                con.ok('You are using the latest version of mole');
            }
        }
    });
}

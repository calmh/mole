"use strict";

var debuggable = require('debuggable');
var libversion = require('version');

var con = require('../lib/console');

module.exports = version ;
version.help = 'Check repository for updates';
version.prio = 8;
debuggable(version);

// Fetch the latest version number for mole from the npm repository and print a
// 'time to upgrade'-message if there's a mismatch.

function version(opts, state) {
    con.info('Client: mole v' + state.pkg.version);

    libversion.fetch('mole', function (err, ver) {
        con.info('Latest: mole v' + ver);
        if (!err && ver) {
            if (ver !== state.pkg.version) {
                con.info('Use "sudo npm -g update mole" to upgrade');
            } else {
                con.ok('You are using the latest version of mole');
            }
        }
    });

    state.client.getPkg(function (pkg) {
        con.info('Server: mole v' + pkg.version);
    });
}
